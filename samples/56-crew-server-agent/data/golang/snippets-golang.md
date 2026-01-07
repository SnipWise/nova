## Hello World
Basic program structure and main function
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

----------

## Variable Declaration
Different ways to declare and initialize variables
```go
package main

import "fmt"

func main() {
    var name string = "John"    // Explicit type
    var age = 30               // Type inference
    city := "New York"         // Short declaration
    var x, y int = 1, 2       // Multiple variables
    
    fmt.Printf("Name: %s, Age: %d\n", name, age)
}
```

----------

## Constants
Declaring and using constants
```go
package main

import "fmt"

const (
    Pi       = 3.14159
    MaxUsers = 100
)

func main() {
    fmt.Printf("Pi: %f, Max Users: %d\n", Pi, MaxUsers)
}
```

----------

## Arrays and Slices
Working with arrays and slices
```go
package main

import "fmt"

func main() {
    arr := [3]int{1, 2, 3}
    slice := []string{"apple", "banana", "cherry"}
    slice = append(slice, "date")
    
    fmt.Printf("Array: %v\n", arr)
    fmt.Printf("Slice: %v, Len: %d\n", slice, len(slice))
}
```

----------

## Maps
Creating and manipulating maps
```go
package main

import "fmt"

func main() {
    fruits := map[string]int{"orange": 10, "grape": 15}
    fruits["apple"] = 5
    
    value, exists := fruits["apple"]
    fmt.Printf("Apple exists: %t, value: %d\n", exists, value)
    
    for key, value := range fruits {
        fmt.Printf("%s: %d\n", key, value)
    }
}
```

----------

## For Loops
Different types of for loops
```go
package main

import "fmt"

func main() {
    // Basic for loop
    for i := 0; i < 3; i++ {
        fmt.Printf("i = %d\n", i)
    }
    
    // Range over slice
    names := []string{"Alice", "Bob"}
    for index, name := range names {
        fmt.Printf("%d: %s\n", index, name)
    }
}
```

----------

## If-Else Statements
Conditional statements
```go
package main

import "fmt"

func main() {
    age := 25
    if age >= 18 {
        fmt.Println("Adult")
    } else {
        fmt.Println("Minor")
    }
    
    // If with short variable declaration
    if num := 42; num > 40 {
        fmt.Printf("Number %d > 40\n", num)
    }
}
```

----------

## Switch Statements
Switch case statements
```go
package main

import "fmt"

func main() {
    day := 3
    switch day {
    case 1:
        fmt.Println("Monday")
    case 2:
        fmt.Println("Tuesday")
    case 3:
        fmt.Println("Wednesday")
    default:
        fmt.Println("Other")
    }
}
```

----------

## Functions
Function declaration and return values
```go
package main

import "fmt"

func add(a, b int) int {
    return a + b
}

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

func main() {
    sum := add(5, 3)
    result, err := divide(10, 2)
    fmt.Printf("Sum: %d, Division: %f, Error: %v\n", sum, result, err)
}
```

----------

## Structs
Defining and using structs
```go
package main

import "fmt"

type Person struct {
    Name string
    Age  int
}

func (p Person) Greet() string {
    return fmt.Sprintf("Hello, I'm %s", p.Name)
}

func main() {
    person := Person{Name: "Alice", Age: 30}
    fmt.Printf("Person: %+v\n", person)
    fmt.Println(person.Greet())
}
```

----------

## Interfaces
Defining and implementing interfaces
```go
package main

import "fmt"

type Shape interface {
    Area() float64
}

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func main() {
    var s Shape = Rectangle{Width: 5, Height: 3}
    fmt.Printf("Area: %.2f\n", s.Area())
}
```

----------

## Error Handling
Proper error handling patterns
```go
package main

import (
    "errors"
    "fmt"
)

var ErrDivisionByZero = errors.New("division by zero")

func safeDivide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, ErrDivisionByZero
    }
    return a / b, nil
}

func main() {
    result, err := safeDivide(10, 0)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Printf("Result: %f\n", result)
    }
}
```

----------

## Pointers
Working with pointers
```go
package main

import "fmt"

func modifyValue(x *int) {
    *x = 100
}

func main() {
    value := 42
    ptr := &value
    
    fmt.Printf("Value: %d, Address: %p\n", value, ptr)
    fmt.Printf("Value via pointer: %d\n", *ptr)
    
    modifyValue(&value)
    fmt.Printf("Modified value: %d\n", value)
}
```

----------

## Goroutines
Concurrent programming with goroutines
```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Printf("Worker %d starting\n", id)
    time.Sleep(100 * time.Millisecond)
    fmt.Printf("Worker %d done\n", id)
}

func main() {
    var wg sync.WaitGroup
    for i := 1; i <= 3; i++ {
        wg.Add(1)
        go worker(i, &wg)
    }
    wg.Wait()
}
```

----------

## Channels
Communication between goroutines
```go
package main

import "fmt"

func producer(ch chan<- int) {
    for i := 1; i <= 3; i++ {
        ch <- i
        fmt.Printf("Sent: %d\n", i)
    }
    close(ch)
}

func main() {
    ch := make(chan int)
    go producer(ch)
    
    for value := range ch {
        fmt.Printf("Received: %d\n", value)
    }
}
```

----------

## File I/O
Reading and writing files
```go
package main

import (
    "fmt"
    "io/ioutil"
    "os"
)

func main() {
    content := "Hello, World!\n"
    err := ioutil.WriteFile("test.txt", []byte(content), 0644)
    if err != nil {
        fmt.Printf("Error writing: %v\n", err)
        return
    }
    
    data, err := ioutil.ReadFile("test.txt")
    if err != nil {
        fmt.Printf("Error reading: %v\n", err)
        return
    }
    
    fmt.Printf("File content: %s", string(data))
    os.Remove("test.txt")
}
```

----------

## JSON Handling
Marshal and unmarshal JSON data
```go
package main

import (
    "encoding/json"
    "fmt"
)

type Person struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email,omitempty"`
}

func main() {
    person := Person{Name: "Alice", Age: 30, Email: "alice@example.com"}
    
    jsonData, _ := json.Marshal(person)
    fmt.Printf("JSON: %s\n", string(jsonData))
    
    var newPerson Person
    json.Unmarshal(jsonData, &newPerson)
    fmt.Printf("Unmarshaled: %+v\n", newPerson)
}
```

----------

## HTTP Client
Making HTTP requests
```go
package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
)

func main() {
    resp, err := http.Get("https://httpbin.org/json")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Error reading response: %v\n", err)
        return
    }
    
    fmt.Printf("Response: %s\n", string(body))
}
```

----------

## HTTP Server
Creating a simple HTTP server
```go
package main

import (
    "fmt"
    "log"
    "net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"message": "Hello, JSON!"}`)
}

func main() {
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/json", jsonHandler)
    
    fmt.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

----------

## String Manipulation
Common string operations
```go
package main

import (
    "fmt"
    "strconv"
    "strings"
)

func main() {
    text := "  Hello, World!  "
    
    fmt.Printf("Trimmed: '%s'\n", strings.TrimSpace(text))
    fmt.Printf("Upper: %s\n", strings.ToUpper(text))
    fmt.Printf("Contains 'World': %t\n", strings.Contains(text, "World"))
    
    words := strings.Fields(strings.TrimSpace(text))
    fmt.Printf("Words: %v\n", words)
    
    // String conversion
    num, _ := strconv.Atoi("123")
    fmt.Printf("String to int: %d\n", num)
}
```

----------

## Time Operations
Working with time and date
```go
package main

import (
    "fmt"
    "time"
)

func main() {
    now := time.Now()
    fmt.Printf("Current time: %v\n", now)
    fmt.Printf("Formatted: %s\n", now.Format("2006-01-02 15:04:05"))
    
    tomorrow := now.Add(24 * time.Hour)
    fmt.Printf("Tomorrow: %v\n", tomorrow)
    
    // Parse time string
    timeStr := "2023-12-25 10:30:00"
    parsed, _ := time.Parse("2006-01-02 15:04:05", timeStr)
    fmt.Printf("Parsed: %v\n", parsed)
}
```

----------

## Testing
Unit tests and table-driven tests
```go
// math.go
package main

func Add(a, b int) int {
    return a + b
}

// math_test.go
package main

import "testing"

func TestAdd(t *testing.T) {
    tests := []struct {
        a, b, expected int
    }{
        {2, 3, 5},
        {0, 0, 0},
        {-1, 1, 0},
    }
    
    for _, test := range tests {
        result := Add(test.a, test.b)
        if result != test.expected {
            t.Errorf("Add(%d, %d) = %d; want %d", 
                test.a, test.b, result, test.expected)
        }
    }
}
```

----------

## Context Usage
Using context for cancellation and timeouts
```go
package main

import (
    "context"
    "fmt"
    "time"
)

func longTask(ctx context.Context) error {
    select {
    case <-time.After(2 * time.Second):
        fmt.Println("Task completed")
        return nil
    case <-ctx.Done():
        fmt.Printf("Task cancelled: %v\n", ctx.Err())
        return ctx.Err()
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    longTask(ctx)
}
```

----------

## Regular Expressions
Pattern matching with regex
```go
package main

import (
    "fmt"
    "regexp"
)

func main() {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    
    emails := []string{"user@example.com", "invalid.email", "test@domain.co.uk"}
    
    for _, email := range emails {
        isValid := emailRegex.MatchString(email)
        fmt.Printf("%s: %t\n", email, isValid)
    }
    
    text := "Contact us at info@company.com"
    matches := emailRegex.FindAllString(text, -1)
    fmt.Printf("Found emails: %v\n", matches)
}
```

----------

## Command Line Arguments
Parsing command line arguments
```go
package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    var name = flag.String("name", "World", "Name to greet")
    var age = flag.Int("age", 0, "Age of person")
    var verbose = flag.Bool("verbose", false, "Verbose output")
    
    flag.Parse()
    
    if *verbose {
        fmt.Printf("Name: %s, Age: %d\n", *name, *age)
    }
    
    fmt.Printf("Hello, %s!\n", *name)
    
    // Remaining arguments
    args := flag.Args()
    if len(args) > 0 {
        fmt.Printf("Extra args: %v\n", args)
    }
}
```

----------

## Database Connection
Basic database operations with PostgreSQL
```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    
    _ "github.com/lib/pq"
)

type User struct {
    ID    int
    Name  string
    Email string
}

func main() {
    db, err := sql.Open("postgres", "host=localhost dbname=test sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    var user User
    err = db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", 1).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Printf("User: %+v\n", user)
    }
}
```

----------

## Logging
Different logging approaches
```go
package main

import (
    "log"
    "log/slog"
    "os"
)

func main() {
    // Standard logging
    log.Println("Standard log message")
    log.Printf("User %s logged in", "Alice")
    
    // Custom logger
    logger := log.New(os.Stdout, "APP: ", log.LstdFlags)
    logger.Println("Custom logger message")
    
    // Structured logging (Go 1.21+)
    slog.Info("User action", 
        "user_id", 123, 
        "action", "login")
    
    slog.Error("Database error", 
        "error", "connection timeout",
        "retry_count", 3)
}
```

----------

## Configuration Management
Managing application configuration
```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
)

type Config struct {
    Server struct {
        Host string `json:"host"`
        Port int    `json:"port"`
    } `json:"server"`
    Database struct {
        Host     string `json:"host"`
        Username string `json:"username"`
    } `json:"database"`
}

func loadConfig(filename string) (*Config, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    var config Config
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&config)
    return &config, err
}

func main() {
    config, err := loadConfig("config.json")
    if err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        return
    }
    
    fmt.Printf("Server: %s:%d\n", config.Server.Host, config.Server.Port)
}
```

----------

## Middleware Pattern
HTTP middleware implementation
```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        log.Printf("Started %s %s", r.Method, r.URL.Path)
        next(w, r)
        log.Printf("Completed in %v", time.Since(start))
    }
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func main() {
    http.HandleFunc("/", loggingMiddleware(homeHandler))
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

----------

## Worker Pool Pattern
Concurrent task processing with worker pool
```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type Job struct {
    ID   int
    Data string
}

func worker(id int, jobs <-chan Job, results chan<- string, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job.ID)
        time.Sleep(100 * time.Millisecond)
        results <- fmt.Sprintf("Job %d completed by worker %d", job.ID, id)
    }
}

func main() {
    jobs := make(chan Job, 5)
    results := make(chan string, 5)
    
    var wg sync.WaitGroup
    
    // Start workers
    for i := 1; i <= 3; i++ {
        wg.Add(1)
        go worker(i, jobs, results, &wg)
    }
    
    // Send jobs
    for i := 1; i <= 5; i++ {
        jobs <- Job{ID: i, Data: fmt.Sprintf("task-%d", i)}
    }
    close(jobs)
    
    // Collect results
    go func() {
        wg.Wait()
        close(results)
    }()
    
    for result := range results {
        fmt.Println(result)
    }
}
```