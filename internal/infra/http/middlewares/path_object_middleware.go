package middlewares

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/BohdanBoriak/boilerplate-go-back/internal/domain" // Додано для domain.User та domain.Task
	"github.com/BohdanBoriak/boilerplate-go-back/internal/infra/http/controllers"
	"github.com/go-chi/chi/v5"
	"github.com/upper/db/v4"
)

type Findable interface {
	Find(uint64) (interface{}, error)
}

func PathObject(pathKey string, ctxKey controllers.CtxKey, service Findable) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			id, err := strconv.ParseUint(chi.URLParam(r, pathKey), 10, 64)
			if err != nil {
				err = fmt.Errorf("invalid %s parameter(only non-negative integers)", pathKey)
				log.Print(err)
				controllers.BadRequest(w, err)
				return
			}

			obj, err := service.Find(id)
			if err != nil {
				log.Print(err)
				errInt4 := fmt.Errorf("%d is greater than maximum value for Int4", id)
				if err == db.ErrNoMoreRows || err.Error() == errInt4.Error() {
					err = fmt.Errorf("record not found")
					controllers.NotFound(w, err)
					return
				}
				controllers.InternalServerError(w, err)
				return
			}

			// перевірка власника таску
			if ctxKey == controllers.TaskKey {
				task, okTask := obj.(domain.Task)
				user, okUser := r.Context().Value(controllers.UserKey).(domain.User)

				if !okTask {
					log.Print("PathObjectMiddleware")
					controllers.InternalServerError(w, errors.New("server error"))
					return
				}
				if !okUser {
					log.Print("PathObjectMiddleware")
					controllers.Unauthorized(w, errors.New("user not found"))
					return
				}

				// чи однакові ID користувача в таску співпадає з ID поточного користувача
				if task.UserId != user.Id {
					controllers.Forbidden(w, errors.New("access denied"))
					return
				}
			}

			ctx := context.WithValue(r.Context(), ctxKey, obj)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(hfn)
	}
}
