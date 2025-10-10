# Patch documentation
The patch in `strict-gin.tmpl` is the only difference between this fork and upstream.

## Motivation

Moment's [Error types](https://docs.moment.com/reference/api-errors) require that we send back a JSON like

```json
{
	"error": "Validation Error.",
	"error_type": "invalid_request"
}
```

However, the short-circuiting error handling in `strict-server` does not support this.

## Patch

The patch contained in `patch.patch`.
