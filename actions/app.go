package actions

import (
	"net/http"
	"sync"

	"todo/models"
	"todo/public"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v3/pop/popmw"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/middleware/csrf"
	"github.com/gobuffalo/middleware/forcessl"
	"github.com/gobuffalo/middleware/paramlogger"
	"github.com/unrolled/secure"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")

var (
	app     *buffalo.App
	appOnce sync.Once
)

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	appOnce.Do(func() {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_todo_session",
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//   c.Value("tx").(*pop.Connection)
		app.Use(popmw.Transaction(models.DB))

		app.GET("/", TasksIndex)
		app.POST("/tasks", TasksCreate)
		app.GET("/tasks/{task_id}/edit", TasksEdit)
		app.PUT("/tasks/{task_id}", TasksUpdate)
		app.PUT("/tasks/{task_id}/toggle", TasksToggle)
		app.DELETE("/tasks/{task_id}", TasksDestroy)
		app.GET("/healthz", HealthCheck)

		app.ServeFiles("/", http.FS(public.FS())) // serve files from the public directory
	})

	return app
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
