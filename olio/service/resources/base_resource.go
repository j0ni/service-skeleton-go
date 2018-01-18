package resources

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/jsonapi"
	"github.com/j0ni/service-skeleton-go/olio/api"
	"github.com/j0ni/service-skeleton-go/olio/util"
)

type BaseResource struct {
	ginEngine *gin.Engine
}

func (self *BaseResource) ReturnJSON(c *gin.Context, status int, record interface{}) {
	w := c.Writer

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	if record != nil {
		c.IndentedJSON(status, record)
	}
}

func (self *BaseResource) ReturnJSONAPI(c *gin.Context, status int, record interface{}) {
	w := c.Writer

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	if record != nil {
		if err := jsonapi.MarshalPayload(w, record); err != nil {
			self.ReturnJSONException(c, api.NewRuntimeException(err.Error()))
		}
	}
}

func (self *BaseResource) ReturnJSONAPIArray(c *gin.Context, status int, records []interface{}) {
	w := c.Writer

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	if len(records) < 1 {
		c.JSON(status, jsonapi.ManyPayload{
			Data: []*jsonapi.Node{},
		})
	} else {
		if err := jsonapi.MarshalPayload(w, records); err != nil {
			self.ReturnJSONException(c, api.NewRuntimeException(err.Error()))
		}
	}
}

func (self *BaseResource) ReturnXML(c *gin.Context, status int, record interface{}) {
	w := c.Writer

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/xml")

	if record != nil {
		output, err := xml.Marshal(record)
		if err == nil {
			fmt.Println(string(output))
			c.Writer.Write(output)
		}
	}
}

func (self *BaseResource) ReturnError(c *gin.Context, status int, message string) {
	var outboundStruct struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}

	outboundStruct.Error = message
	outboundStruct.Code = status
	self.ReturnJSON(c, status, outboundStruct)
}

func (self *BaseResource) ReturnJSONException(c *gin.Context, exception *api.Exception) {
	self.ReturnError(c, exception.ErrorCode, exception.Err)
}

func (self *BaseResource) ReturnXMLError(c *gin.Context, status int, message string) {
	type Error struct {
		Code    int    `json:"code"`
		Message string `json:"error"`
	}

	outboundStruct := Error{}
	outboundStruct.Message = message
	outboundStruct.Code = status
	self.ReturnXML(c, status, outboundStruct)
}

func (self *BaseResource) ParseString(c *gin.Context, paramName string) string {
	param := c.Query(paramName)
	if param == "" {
		param = c.Param(paramName)
	}

	return param
}

func (self *BaseResource) ParseInt(c *gin.Context, paramName string) int64 {
	param := self.ParseString(c, paramName)
	if param == "" {
		return 0
	}

	return util.StringToInt64(param)
}

func (self *BaseResource) ParseDate(c *gin.Context, field string) *time.Time {
	if field == "" {
		return nil
	}
	param := c.Param(field)
	if param == "" {
		// try parameter
		param = c.Query(field)
	}

	if param == "" {
		return nil
	}

	if _, err := strconv.Atoi(param); err == nil {
		t, err := msToTime(param)
		if err != nil {
			return nil
		}
		return &t
	}

	startAt, e := time.Parse(time.RFC3339, param)
	startAt = startAt.UTC()
	if e != nil {
		return nil
	}
	return &startAt
}

func (self *BaseResource) ParseArray(c *gin.Context, paramName string) []string {
	param := self.ParseString(c, paramName)
	if param == "" {
		return nil
	}
	return strings.Split(param, ",")
}

func (self *BaseResource) ParseJsonapi(c *gin.Context, record interface{}) error {
	if err := jsonapi.UnmarshalPayload(c.Request.Body, record); err != nil {
		return errors.New("Problem unmarshalling incoming payload [" + err.Error() + "]")
	}

	return nil
}

func (self *BaseResource) ParseJson(c *gin.Context, record interface{}) error {
	if err := c.BindJSON(&record); err != nil {
		return errors.New("Problem unmarshalling incoming payload [" + err.Error() + "]")
	}

	return nil
}

func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

func (self *BaseResource) SetPaginationHeaders(c *gin.Context, limit int64, offset int64, totalCount int64) {
	c.Header("X-Total-Count", strconv.FormatInt(totalCount, 10))
	setLinks(c, limit, offset, totalCount)
}

func (self *BaseResource) HasPaginationParams(c *gin.Context) bool {
	return self.ParseString(c, "offset") != "" && self.ParseString(c, "limit") != ""
}

func GetBaseUrl(c *gin.Context) string {
	var proto string

	// assume if port is specified, request is not HTTPS, otherwise it is
	if strings.Index(c.Request.Host, ":") > -1 {
		proto = "http://"
	} else {
		proto = "https://"
	}

	return proto + c.Request.Host + c.Request.URL.String()
}

func setLinks(c *gin.Context, limit int64, offset int64, totalCount int64) {
	baseUrl := GetBaseUrl(c)

	lastOffset := totalCount - totalCount%limit
	nextOffset := offset + limit
	prevOffset := offset - limit

	var links []string

	if offset < lastOffset {
		links = append(links, buildLink(baseUrl, "next", limit, nextOffset))
		links = append(links, buildLink(baseUrl, "last", limit, lastOffset))
	}

	if offset > 1 {
		links = append(links, buildLink(baseUrl, "first", limit, 0))
		links = append(links, buildLink(baseUrl, "prev", limit, prevOffset))
	}

	linkStr := strings.Join(links, ", ")
	c.Header("Link", linkStr)
}

func buildLink(baseUrl string, linkType string, limit, offset int64) string {
	newUrl, _ := url.Parse(baseUrl)

	query := newUrl.Query()
	query.Set("limit", strconv.FormatInt(limit, 10))
	query.Set("offset", strconv.FormatInt(offset, 10))
	newUrl.RawQuery = query.Encode()

	link := "<" + newUrl.String() + `>; rel="` + linkType + `"`
	return link
}
