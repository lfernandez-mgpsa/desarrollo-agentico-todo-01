package actions

import (
	"net/http"
	"strings"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

// TasksIndex muestra la lista de tareas y el formulario de creación.
func TasksIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	tasks := models.Tasks{}
	if err := tx.Order("created_at asc, id asc").All(&tasks); err != nil {
		return err
	}

	c.Set("tasks", tasks)
	c.Set("task", &models.Task{})
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("tasks/index.plush.html"))
}

// TasksCreate crea una tarea nueva a partir del formulario.
func TasksCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := c.Bind(task); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(task.Title)

	verrs, err := tx.ValidateAndCreate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		tasks := models.Tasks{}
		if err := tx.Order("created_at asc, id asc").All(&tasks); err != nil {
			return err
		}

		c.Set("tasks", tasks)
		c.Set("task", task)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tasks/index.plush.html"))
	}

	c.Flash().Add("success", "Tarea creada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksEdit muestra el formulario de edición del título de una tarea pendiente.
func TasksEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if task.Completed {
		return redirectCompletedNotEditable(c)
	}

	c.Set("task", task)
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("tasks/edit.plush.html"))
}

// TasksUpdate actualiza el título de una tarea pendiente.
func TasksUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if task.Completed {
		return redirectCompletedNotEditable(c)
	}

	// Se enlaza sobre una tarea vacía para que el formulario solo pueda cambiar el título.
	form := &models.Task{}
	if err := c.Bind(form); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(form.Title)

	verrs, err := tx.ValidateAndUpdate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		c.Set("task", task)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tasks/edit.plush.html"))
	}

	c.Flash().Add("success", "Tarea actualizada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// redirectCompletedNotEditable vuelve a la lista avisando de que una tarea completada no se edita.
func redirectCompletedNotEditable(c buffalo.Context) error {
	c.Flash().Add("error", "No se puede editar una tarea completada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksToggle alterna una tarea entre completada y pendiente.
func TasksToggle(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	task.Completed = !task.Completed
	if err := tx.Update(task); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksDestroy elimina una tarea.
func TasksDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(task); err != nil {
		return err
	}

	c.Flash().Add("success", "Tarea eliminada.")
	return c.Redirect(http.StatusSeeOther, "/")
}
