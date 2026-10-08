package actions

import (
	"fmt"
	"net/url"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_TasksIndex_EmptyState() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay tareas todavía")
}

func (as *ActionSuite) Test_TasksIndex_ListsTasks() {
	as.LoadFixture("tres tareas")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Comprar material de oficina")
	as.Contains(body, "Llamar al gestor")
	as.Contains(body, "Publicar la oferta de empleo")
	as.Contains(body, "Completada")
}

func (as *ActionSuite) Test_TasksCreate_Valid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "  Nueva tarea de prueba  "})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Nueva tarea de prueba", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksCreate_EmptyTitle_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: ""})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_OnlySpaces_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "     "})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksToggle_CompletesAndReopens() {
	task := &models.Task{Title: "Pendiente de completar"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)

	res = as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/toggle", id).Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksToggle_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/toggle").Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksDestroy() {
	task := &models.Task{Title: "Tarea que se elimina"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Delete()
	as.Equal(303, res.Code)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksDestroy_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Delete()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksIndex_EditLinkOnlyForPending() {
	pending := &models.Task{Title: "Tarea pendiente"}
	verrs, err := as.DB.ValidateAndCreate(pending)
	as.NoError(err)
	as.False(verrs.HasAny())

	completed := &models.Task{Title: "Tarea completada", Completed: true}
	verrs, err = as.DB.ValidateAndCreate(completed)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, fmt.Sprintf("/tasks/%s/edit", pending.ID))
	as.NotContains(body, fmt.Sprintf("/tasks/%s/edit", completed.ID))
}

func (as *ActionSuite) Test_TasksEdit_ShowsForm() {
	task := &models.Task{Title: "Título original"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Editar tarea")
	as.Contains(res.Body.String(), "Título original")
}

func (as *ActionSuite) Test_TasksEdit_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/edit", id).Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_Completed_Redirects() {
	task := &models.Task{Title: "Ya completada", Completed: true}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))
}

func (as *ActionSuite) Test_TasksUpdate_Valid() {
	task := &models.Task{Title: "Título con errata"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Put(&models.Task{Title: "  Título corregido  "})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	as.NoError(as.DB.Reload(task))
	as.Equal("Título corregido", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_EmptyTitle_Invalid() {
	task := &models.Task{Title: "Título que se mantiene"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Put(&models.Task{Title: "   "})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	as.NoError(as.DB.Reload(task))
	as.Equal("Título que se mantiene", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_TooLong_Invalid() {
	task := &models.Task{Title: "Título que se mantiene"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Put(&models.Task{Title: strings.Repeat("a", models.TitleMaxLength+1)})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título no puede superar")

	as.NoError(as.DB.Reload(task))
	as.Equal("Título que se mantiene", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_Completed_Rejected() {
	task := &models.Task{Title: "Ya completada", Completed: true}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Put(&models.Task{Title: "Otro título"})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	as.NoError(as.DB.Reload(task))
	as.Equal("Ya completada", task.Title)
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Put(&models.Task{Title: "Da igual"})
	as.Equal(404, res.Code)
}
