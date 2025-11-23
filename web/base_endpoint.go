package web

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-yaaf/yaaf-common/entity"
	"github.com/go-yaaf/yaaf-common/utils/collections"

	. "github.com/go-yaaf/yaaf-common-net/model"
	"github.com/go-yaaf/yaaf-common-net/utils"
)

const (
	UNKNOWN int = 0
	APIKEY      = 1
	TOKEN       = 3
)

type BaseEndPoint struct{}

// GetTokenData extract security token data from Authorization header
func (b *BaseEndPoint) GetTokenData(c *gin.Context) *TokenData {
	token := c.GetHeader("X-ACCESS-TOKEN")
	if td, err := utils.TokenUtils().ParseToken(token); err != nil {
		_ = c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid access token"))
		return nil
	} else {
		return td
	}
}

// GetTimezoneOffset returns the value of timezone offset header in minutes
func (b *BaseEndPoint) GetTimezoneOffset(c *gin.Context) int {

	now := time.Now()
	_, offsetSeconds := now.Zone()
	localOffset := offsetSeconds / 60

	offset := c.GetHeader("X-TIMEZONE-OFFSET")
	clientOffset, err := strconv.Atoi(offset)
	if err != nil {
		return 0
	} else {
		return localOffset + clientOffset
	}
}

// ResolveRemoteIp extract remote ip from HTTP header X-Forwarded-For
func (b *BaseEndPoint) ResolveRemoteIp(c *gin.Context) (ip string) {
	if ip = c.GetHeader("X-Forwarded-For"); len(ip) == 0 {
		ip = c.RemoteIP()
	}
	return
}

// GetParamAsString extract parameter value from query string as string
func (b *BaseEndPoint) GetParamAsString(c *gin.Context, paramName string, defaultValue string) string {
	// First try query params, then try path param, then return default
	if str, ok := c.GetQuery(paramName); ok {
		return strings.Trim(str, " ")
	} else if str, ok = c.Params.Get(paramName); ok {
		return strings.Trim(str, " ")
	} else {
		return defaultValue
	}
}

// GetParamAsInt extract parameter value from query string as int
func (b *BaseEndPoint) GetParamAsInt(c *gin.Context, paramName string, defaultValue int) int {

	param := ""
	// First try query params, then try path param, then return default
	if str, ok := c.GetQuery(paramName); ok {
		param = str
	} else if str, ok = c.Params.Get(paramName); ok {
		param = str
	} else {
		return defaultValue
	}

	if res, err := strconv.Atoi(param); err != nil {
		return defaultValue
	} else {
		return res
	}
}

// GetParamAsFloat extract parameter value from query string as float64
func (b *BaseEndPoint) GetParamAsFloat(c *gin.Context, paramName string, defaultValue float64) float64 {

	param := ""
	// First try query params, then try path param, then return default
	if str, ok := c.GetQuery(paramName); ok {
		param = str
	} else if str, ok = c.Params.Get(paramName); ok {
		param = str
	} else {
		return defaultValue
	}

	if res, err := strconv.ParseFloat(param, 32); err != nil {
		return defaultValue
	} else {
		return res
	}
}

// GetParamAsStringArray extract parameter array values from query string
// This supports multiple values query string e.g. https//some/domain?id=str1&id=str2&id=str3
func (b *BaseEndPoint) GetParamAsStringArray(c *gin.Context, paramName string) (res []string) {
	list, _ := c.GetQueryArray(paramName)
	for _, item := range list {
		if len(item) > 0 {
			items := strings.Split(item, ",")
			res = append(res, items...)
		}
	}
	return
}

// GetParamAsIntArray extract parameter array values from query string
// This supports multiple values query string e.g. https//some/domain?id=1&id=2&id=3
func (b *BaseEndPoint) GetParamAsIntArray(c *gin.Context, paramName string) (res []int) {
	if params, ok := c.GetQueryArray(paramName); !ok {
		return
	} else {
		for _, p := range params {
			arr := strings.Split(p, ",")
			for _, str := range arr {
				if num, err := strconv.Atoi(str); err == nil {
					res = append(res, num)
				}
			}
		}
	}
	return
}

// GetStringParamArray extract parameter array values from query string
// This supports multiple values query string e.g. https//some/domain?id=1&id=2&id=3
func (b *BaseEndPoint) getStringParamArray(c *gin.Context, paramName string) []string {
	params, _ := c.GetQueryArray(paramName)
	return params
}

// GetParamAsBool extract parameter value from query string as bool
func (b *BaseEndPoint) GetParamAsBool(c *gin.Context, paramName string, defaultValue bool) bool {
	if str, ok := c.GetQuery(paramName); !ok {
		return defaultValue
	} else {
		if res, err := strconv.ParseBool(str); err != nil {
			return defaultValue
		} else {
			return res
		}
	}
}

// GetParamAsEnum extract parameter value from query string as enum
// Enum value can be passed as it's int or string value
func (b *BaseEndPoint) GetParamAsEnum(c *gin.Context, paramName string, enum any, defaultValue int) int {

	param := ""
	// First try query params, then try path param, then return default
	if str, ok := c.GetQuery(paramName); ok {
		param = str
	} else if str, ok = c.Params.Get(paramName); ok {
		param = str
	} else {
		return defaultValue
	}

	if res, err := strconv.Atoi(param); err != nil {
		// try enum string
		if n, e := getEnumValueFromName(enum, param); e != nil {
			return defaultValue
		} else {
			return n
		}
	} else {
		return res
	}
}

// GetParamAsEnumArray Extract parameter array values from query string as enums
// This supports multiple values query string e.g. https//some/domain?id=1&id=2&id=3
func (b *BaseEndPoint) GetParamAsEnumArray(c *gin.Context, paramName string, enum interface{}) (res []int) {
	params, _ := c.GetQueryArray(paramName)
	for _, p := range params {
		arr := strings.Split(p, ",")

		for _, str := range arr {
			if len(str) == 0 {
				continue
			}
			if num, err := strconv.Atoi(str); err != nil {
				// try enum string
				if n, e := getEnumValueFromName(enum, str); e == nil {
					res = append(res, n)
				}
			} else {
				res = append(res, num)
			}
		}
	}
	return
}

// GetParamAsTimestamp extract parameter value from query string as Timestamp
// Timestamp can be absolute (epoch milliseconds) or relative (delta from now)
func (b *BaseEndPoint) GetParamAsTimestamp(c *gin.Context, paramName string, defaultValue int) entity.Timestamp {

	param := ""
	// First try query params, then try path param, then return default
	if str, ok := c.GetQuery(paramName); ok {
		param = str
	} else if str, ok = c.Params.Get(paramName); ok {
		param = str
	} else {
		return relativeToAbsoluteTime(entity.Timestamp(defaultValue))
	}
	if res, err := strconv.Atoi(param); err != nil {
		return relativeToAbsoluteTime(entity.Timestamp(defaultValue))
	} else {
		// convert relative to absolute time
		return relativeToAbsoluteTime(entity.Timestamp(res))
	}
}

// Get enum int value from enum name
func getEnumValueFromName(enum any, name string) (result int, err error) {

	ref := reflect.ValueOf(enum)
	typ := ref.Type()

	for i := 0; i < ref.NumField(); i++ {

		s := typ.Field(i).Name
		res := int(ref.Field(i).Int())

		if s == name {
			return res, nil
		}
	}

	return 0, fmt.Errorf("not found")
}

// Get enum int value from enum name
func getEnumValuesFromNames(enum any, values []string) []int {

	// Handle any Panic error
	//defer utils.RecoverAll(func(err any) {
	//	return
	//})

	result := make([]int, 0)

	ref := reflect.ValueOf(enum)
	typ := ref.Type()

	for i := 0; i < ref.NumField(); i++ {

		s := typ.Field(i).Name
		res := int(ref.Field(i).Int())

		if collections.Include(values, s) {
			result = append(result, res)
		}
	}

	return result
}

// Convert relative time (0 or negative number) to absolute time
// If time is provided as 0 or negative number, it is considered as relative time in milliseconds
func relativeToAbsoluteTime(input entity.Timestamp) entity.Timestamp {
	if input < 0 {
		return entity.Now() + input
	} else {
		return input
	}
}

// GetPathParamAsString extract parameter value from path as string
func (b *BaseEndPoint) GetPathParamAsString(c *gin.Context, paramName string, defaultValue string) string {
	if val, ok := c.Params.Get(paramName); ok {
		return val
	} else {
		return defaultValue
	}
}

// GetPathParamAsInt extract parameter value from path as int
func (b *BaseEndPoint) GetPathParamAsInt(c *gin.Context, paramName string, defaultValue int) int {
	if val, ok := c.Params.Get(paramName); ok {
		if res, err := strconv.Atoi(val); err != nil {
			return defaultValue
		} else {
			return res
		}
	} else {
		return defaultValue
	}
}

// GetPathParamAsEnum extract parameter value from path string as enum
// Enum value can be passed as it's int or string value
func (b *BaseEndPoint) GetPathParamAsEnum(c *gin.Context, paramName string, enum any, defaultValue int) int {

	if val, ok := c.Params.Get(paramName); !ok {
		return defaultValue
	} else {
		if res, err := strconv.Atoi(val); err != nil {
			if n, e := getEnumValueFromName(enum, val); e != nil {
				return defaultValue
			} else {
				return n
			}
		} else {
			return res
		}
	}
}
