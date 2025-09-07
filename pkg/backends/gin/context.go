package gin

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Context wraps gin.Context with additional functionality
type Context struct {
	ginContext *gin.Context
}

// NewContext creates a new Context wrapper
func NewContext(ginContext *gin.Context) *Context {
	return &Context{
		ginContext: ginContext,
	}
}

// String writes a string response with the given HTTP status code
func (c *Context) String(code int, format string, values ...interface{}) {
	c.ginContext.String(code, format, values...)
}

// JSON serializes the given struct as JSON into the response body
func (c *Context) JSON(code int, obj interface{}) {
	c.ginContext.JSON(code, obj)
}

// XML serializes the given struct as XML into the response body
func (c *Context) XML(code int, obj interface{}) {
	c.ginContext.XML(code, obj)
}

// Data writes raw data into the response body
func (c *Context) Data(code int, contentType string, data []byte) {
	c.ginContext.Data(code, contentType, data)
}

// Request returns the underlying HTTP request
func (c *Context) Request() *http.Request {
	return c.ginContext.Request
}

// Writer returns the response writer
func (c *Context) Writer() gin.ResponseWriter {
	return c.ginContext.Writer
}

// GetHeader returns the value of the request header
func (c *Context) GetHeader(key string) string {
	return c.ginContext.GetHeader(key)
}

// Header sets a response header
func (c *Context) Header(key, value string) {
	c.ginContext.Header(key, value)
}

// Param returns the value of the URL param
func (c *Context) Param(key string) string {
	return c.ginContext.Param(key)
}

// Query returns the keyed url query value if it exists
func (c *Context) Query(key string) string {
	return c.ginContext.Query(key)
}

// DefaultQuery returns the keyed url query value if it exists, otherwise returns the default value
func (c *Context) DefaultQuery(key, defaultValue string) string {
	return c.ginContext.DefaultQuery(key, defaultValue)
}

// PostForm returns the specified key from a POST urlencoded form or multipart form
func (c *Context) PostForm(key string) string {
	return c.ginContext.PostForm(key)
}

// DefaultPostForm returns the specified key from a POST urlencoded form or multipart form,
// otherwise returns the default value
func (c *Context) DefaultPostForm(key, defaultValue string) string {
	return c.ginContext.DefaultPostForm(key, defaultValue)
}

// Bind checks the Content-Type to select a binding engine automatically
func (c *Context) Bind(obj interface{}) error {
	return c.ginContext.Bind(obj)
}

// ShouldBind checks the Content-Type to select a binding engine automatically
func (c *Context) ShouldBind(obj interface{}) error {
	return c.ginContext.ShouldBind(obj)
}

// BindJSON is a shortcut for c.Bind(obj) with binding.JSON
func (c *Context) BindJSON(obj interface{}) error {
	return c.ginContext.BindJSON(obj)
}

// ShouldBindJSON is a shortcut for c.ShouldBind(obj) with binding.JSON
func (c *Context) ShouldBindJSON(obj interface{}) error {
	return c.ginContext.ShouldBindJSON(obj)
}

// Set stores a new key/value pair exclusively for this context
func (c *Context) Set(key string, value interface{}) {
	c.ginContext.Set(key, value)
}

// Get returns the value for the given key, ie: (value, true)
func (c *Context) Get(key string) (value interface{}, exists bool) {
	return c.ginContext.Get(key)
}

// GetString returns the value associated with the key as a string
func (c *Context) GetString(key string) (s string) {
	return c.ginContext.GetString(key)
}

// GetBool returns the value associated with the key as a boolean
func (c *Context) GetBool(key string) (b bool) {
	return c.ginContext.GetBool(key)
}

// GetInt returns the value associated with the key as an integer
func (c *Context) GetInt(key string) (i int) {
	return c.ginContext.GetInt(key)
}

// GetInt64 returns the value associated with the key as an integer
func (c *Context) GetInt64(key string) (i64 int64) {
	return c.ginContext.GetInt64(key)
}

// GetFloat64 returns the value associated with the key as a float64
func (c *Context) GetFloat64(key string) (f64 float64) {
	return c.ginContext.GetFloat64(key)
}

// Abort prevents pending handlers from being called
func (c *Context) Abort() {
	c.ginContext.Abort()
}

// AbortWithStatus calls Abort() and writes the headers with the specified status code
func (c *Context) AbortWithStatus(code int) {
	c.ginContext.AbortWithStatus(code)
}

// AbortWithStatusJSON calls Abort() and then JSON() internally
func (c *Context) AbortWithStatusJSON(code int, jsonObj interface{}) {
	c.ginContext.AbortWithStatusJSON(code, jsonObj)
}

// IsAborted returns true if the current context was aborted
func (c *Context) IsAborted() bool {
	return c.ginContext.IsAborted()
}

// Next should be used only inside middleware
func (c *Context) Next() {
	c.ginContext.Next()
}

// GinContext returns the underlying gin.Context
func (c *Context) GinContext() *gin.Context {
	return c.ginContext
}

// ReadBody reads the request body and returns it as bytes
func (c *Context) ReadBody() ([]byte, error) {
	return io.ReadAll(c.ginContext.Request.Body)
}

// Status sets the HTTP response code
func (c *Context) Status(code int) {
	c.ginContext.Status(code)
}