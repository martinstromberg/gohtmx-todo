package main

import (
    "html/template"
    "github.com/google/uuid"
    "fmt"
    "log"
    "net/http"
    "strings"
)

type Todo struct {
    Id              string
    Title           string
    IsCompleted     bool
}

var todos []*Todo = make([]*Todo, 1)

const BaseTemplate string = "./web/template/base.tmpl"
const TodosIndexTemplate string = "./web/template/todo-index.tmpl"
const TodoItemTemplate string = "./web/template/todo-item.tmpl"

func handler (w http.ResponseWriter, r *http.Request) {
    fmt.Printf("%s %s\n", r.Method, r.URL.Path)

    if len(r.URL.Path) >= 6 && r.URL.Path[1:6] == "todos" {
        handleTodos(w, r)
        return
    }

    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func handleTodos (w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        handleGetTodo(w, r)
        return

    case http.MethodPost:
        handlePostTodo(w, r)
        return
        
    case http.MethodPut:
        handlePutTodo(w, r)
        return

    case http.MethodDelete:
        handleDeleteTodo(w, r)
        return

    }
}

func handleDeleteTodo (w http.ResponseWriter, r *http.Request) {
    idFromUrl := r.URL.Path[7:]

    for index, item := range todos {
        if item.Id != idFromUrl {
            continue
        }

        todos = append(todos[:index], todos[index+1:]...)
        w.WriteHeader(http.StatusOK)
        return
    }

    w.WriteHeader(http.StatusNotFound)
}

func handlePutTodo (w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    idFromUrl := r.URL.Path[7:]
    isCompleted := r.Form.Get("isCompleted")

    for _, item := range todos {
        if item.Id != idFromUrl {
            continue
        }

        item.IsCompleted = isCompleted == "on"

        itemTmpl, err := template.New("todoitem").ParseFiles(TodoItemTemplate)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        itemTmpl.ExecuteTemplate(w, "todoitem", item)

        return
    }

    w.WriteHeader(http.StatusNotFound)
}

func handlePostTodo (w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    title := r.Form.Get("title")

    todo := Todo{
        Id: uuid.New().String(),
        Title: title,
        IsCompleted: false,
    }

    todos = append(todos, &todo)

    itemTmpl, err := template.New("todoitem").ParseFiles(TodoItemTemplate)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    itemTmpl.ExecuteTemplate(w, "todoitem", todo)
}

func handleGetTodo (w http.ResponseWriter, r *http.Request) {
    itemTmpl, err := template.New("todoitem").ParseFiles(TodoItemTemplate)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    strings := make([]template.HTML, len(todos))
    for index, todo := range todos {
        if todo == nil {
            continue
        }

        str, err := renderTodoItem(itemTmpl, *todo)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        
        strings[index] = template.HTML(str)
    }

    ts, err := template.ParseFiles(
        BaseTemplate,
        TodosIndexTemplate,
    )

    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    err = ts.ExecuteTemplate(w, "base", map[string]interface{}{
        "HtmlStrings": strings,
    })
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
}

func renderTodoItem (tmpl *template.Template, item Todo) (string, error) {
    var b strings.Builder

    err := tmpl.ExecuteTemplate(&b, "todoitem", item)

    out := b.String()
    return out, err
}


func main() {
    todos[0] = &Todo{
        Id: uuid.New().String(),
        Title: "Finish the demo",
        IsCompleted: false,
    }

    http.HandleFunc("/", handler)
    http.HandleFunc("/todos", handler)


    log.Fatal(http.ListenAndServe(":8080", nil))
}
