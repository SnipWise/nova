## Singleton
Ensures a single instance exists throughout the application
```go
package main

import (
    "fmt"
    "sync"
)

type singleton struct {
    value string
}

var instance *singleton
var once sync.Once

func GetInstance() *singleton {
    once.Do(func() {
        instance = &singleton{value: "initialized"}
    })
    return instance
}

func main() {
    s1 := GetInstance()
    s2 := GetInstance()
    fmt.Printf("Same instance: %t\n", s1 == s2)
}
```

----------

## Factory
Creates objects without exposing instantiation logic
```go
package main

import "fmt"

type Animal interface {
    Speak() string
}

type Dog struct{}
func (d Dog) Speak() string { return "Woof!" }

type Cat struct{}
func (c Cat) Speak() string { return "Meow!" }

func NewAnimal(animalType string) Animal {
    switch animalType {
    case "dog":
        return Dog{}
    case "cat":
        return Cat{}
    default:
        return nil
    }
}

func main() {
    dog := NewAnimal("dog")
    cat := NewAnimal("cat")
    fmt.Println(dog.Speak(), cat.Speak())
}
```

----------

## Abstract Factory
Creates families of related objects
```go
package main

import "fmt"

type Button interface {
    Render() string
}
type Checkbox interface {
    Check() string
}

type WinButton struct{}
func (w WinButton) Render() string { return "Windows Button" }
type WinCheckbox struct{}
func (w WinCheckbox) Check() string { return "Windows Checkbox" }

type GUIFactory interface {
    CreateButton() Button
    CreateCheckbox() Checkbox
}

type WindowsFactory struct{}
func (w WindowsFactory) CreateButton() Button { return WinButton{} }
func (w WindowsFactory) CreateCheckbox() Checkbox { return WinCheckbox{} }

func main() {
    factory := WindowsFactory{}
    fmt.Println(factory.CreateButton().Render())
}
```

----------

## Builder
Constructs complex objects step by step
```go
package main

import "fmt"

type Server struct {
    Host    string
    Port    int
    Timeout int
    MaxConn int
}

type ServerBuilder struct {
    server Server
}

func NewServerBuilder() *ServerBuilder {
    return &ServerBuilder{}
}

func (b *ServerBuilder) Host(h string) *ServerBuilder {
    b.server.Host = h
    return b
}

func (b *ServerBuilder) Port(p int) *ServerBuilder {
    b.server.Port = p
    return b
}

func (b *ServerBuilder) Build() Server {
    return b.server
}

func main() {
    server := NewServerBuilder().Host("localhost").Port(8080).Build()
    fmt.Printf("%+v\n", server)
}
```

----------

## Functional Options
Flexible configuration using function options
```go
package main

import "fmt"

type Server struct {
    host    string
    port    int
    timeout int
}

type Option func(*Server)

func WithHost(h string) Option {
    return func(s *Server) { s.host = h }
}

func WithPort(p int) Option {
    return func(s *Server) { s.port = p }
}

func WithTimeout(t int) Option {
    return func(s *Server) { s.timeout = t }
}

func NewServer(opts ...Option) *Server {
    s := &Server{host: "localhost", port: 8080, timeout: 30}
    for _, opt := range opts {
        opt(s)
    }
    return s
}

func main() {
    srv := NewServer(WithHost("0.0.0.0"), WithPort(9000))
    fmt.Printf("%+v\n", srv)
}
```

----------

## Prototype
Creates new objects by cloning existing ones
```go
package main

import "fmt"

type Prototype interface {
    Clone() Prototype
}

type Document struct {
    Title   string
    Content string
}

func (d *Document) Clone() Prototype {
    return &Document{
        Title:   d.Title,
        Content: d.Content,
    }
}

func main() {
    original := &Document{Title: "Report", Content: "Data..."}
    clone := original.Clone().(*Document)
    clone.Title = "Report Copy"
    
    fmt.Printf("Original: %s\n", original.Title)
    fmt.Printf("Clone: %s\n", clone.Title)
}
```

----------

## Adapter
Converts interface of a class into another expected interface
```go
package main

import "fmt"

type OldPrinter interface {
    PrintOld(s string) string
}

type LegacyPrinter struct{}

func (l *LegacyPrinter) PrintOld(s string) string {
    return "Legacy: " + s
}

type NewPrinter interface {
    Print(s string) string
}

type PrinterAdapter struct {
    old OldPrinter
}

func (a *PrinterAdapter) Print(s string) string {
    return a.old.PrintOld(s)
}

func main() {
    legacy := &LegacyPrinter{}
    adapter := &PrinterAdapter{old: legacy}
    fmt.Println(adapter.Print("Hello"))
}
```

----------

## Decorator
Adds behavior to objects dynamically
```go
package main

import "fmt"

type Notifier interface {
    Send(msg string) string
}

type EmailNotifier struct{}

func (e *EmailNotifier) Send(msg string) string {
    return "Email: " + msg
}

type SMSDecorator struct {
    wrapped Notifier
}

func (s *SMSDecorator) Send(msg string) string {
    return s.wrapped.Send(msg) + " + SMS: " + msg
}

func main() {
    email := &EmailNotifier{}
    decorated := &SMSDecorator{wrapped: email}
    fmt.Println(decorated.Send("Hello"))
}
```

----------

## Facade
Provides simplified interface to complex subsystem
```go
package main

import "fmt"

type CPU struct{}
func (c *CPU) Start() { fmt.Println("CPU started") }

type Memory struct{}
func (m *Memory) Load() { fmt.Println("Memory loaded") }

type Disk struct{}
func (d *Disk) Read() { fmt.Println("Disk read") }

type ComputerFacade struct {
    cpu    *CPU
    memory *Memory
    disk   *Disk
}

func NewComputer() *ComputerFacade {
    return &ComputerFacade{&CPU{}, &Memory{}, &Disk{}}
}

func (c *ComputerFacade) Start() {
    c.cpu.Start()
    c.memory.Load()
    c.disk.Read()
}

func main() {
    computer := NewComputer()
    computer.Start()
}
```

----------

## Proxy
Controls access to another object
```go
package main

import "fmt"

type Database interface {
    Query(q string) string
}

type RealDatabase struct{}

func (r *RealDatabase) Query(q string) string {
    return "Result for: " + q
}

type DatabaseProxy struct {
    db       *RealDatabase
    isAdmin  bool
}

func (p *DatabaseProxy) Query(q string) string {
    if !p.isAdmin {
        return "Access denied"
    }
    if p.db == nil {
        p.db = &RealDatabase{}
    }
    return p.db.Query(q)
}

func main() {
    proxy := &DatabaseProxy{isAdmin: true}
    fmt.Println(proxy.Query("SELECT *"))
}
```

----------

## Composite
Composes objects into tree structures
```go
package main

import "fmt"

type Component interface {
    GetPrice() float64
}

type Product struct {
    name  string
    price float64
}

func (p *Product) GetPrice() float64 { return p.price }

type Box struct {
    children []Component
}

func (b *Box) Add(c Component) { b.children = append(b.children, c) }

func (b *Box) GetPrice() float64 {
    total := 0.0
    for _, child := range b.children {
        total += child.GetPrice()
    }
    return total
}

func main() {
    box := &Box{}
    box.Add(&Product{"Phone", 500})
    box.Add(&Product{"Charger", 25})
    fmt.Printf("Total: $%.2f\n", box.GetPrice())
}
```

----------

## Bridge
Separates abstraction from implementation
```go
package main

import "fmt"

type Renderer interface {
    RenderCircle(radius float64) string
}

type VectorRenderer struct{}

func (v *VectorRenderer) RenderCircle(r float64) string {
    return fmt.Sprintf("Vector circle r=%.1f", r)
}

type RasterRenderer struct{}

func (r *RasterRenderer) RenderCircle(rad float64) string {
    return fmt.Sprintf("Raster circle r=%.1f", rad)
}

type Circle struct {
    renderer Renderer
    radius   float64
}

func (c *Circle) Draw() string {
    return c.renderer.RenderCircle(c.radius)
}

func main() {
    circle := &Circle{&VectorRenderer{}, 5}
    fmt.Println(circle.Draw())
}
```

----------

## Strategy
Defines family of interchangeable algorithms
```go
package main

import "fmt"

type PaymentStrategy interface {
    Pay(amount float64) string
}

type CreditCard struct{}

func (c *CreditCard) Pay(amount float64) string {
    return fmt.Sprintf("Paid $%.2f via Credit Card", amount)
}

type PayPal struct{}

func (p *PayPal) Pay(amount float64) string {
    return fmt.Sprintf("Paid $%.2f via PayPal", amount)
}

type Checkout struct {
    strategy PaymentStrategy
}

func (c *Checkout) Process(amount float64) string {
    return c.strategy.Pay(amount)
}

func main() {
    checkout := &Checkout{&CreditCard{}}
    fmt.Println(checkout.Process(100.50))
}
```

----------

## Observer
Notifies multiple objects about state changes
```go
package main

import "fmt"

type Observer interface {
    Update(msg string)
}

type Subject struct {
    observers []Observer
}

func (s *Subject) Attach(o Observer) {
    s.observers = append(s.observers, o)
}

func (s *Subject) Notify(msg string) {
    for _, o := range s.observers {
        o.Update(msg)
    }
}

type EmailSubscriber struct{ email string }

func (e *EmailSubscriber) Update(msg string) {
    fmt.Printf("Email to %s: %s\n", e.email, msg)
}

func main() {
    subject := &Subject{}
    subject.Attach(&EmailSubscriber{"user@mail.com"})
    subject.Notify("New event!")
}
```

----------

## Command
Encapsulates request as an object
```go
package main

import "fmt"

type Command interface {
    Execute()
}

type Light struct{}

func (l *Light) On()  { fmt.Println("Light is ON") }
func (l *Light) Off() { fmt.Println("Light is OFF") }

type LightOnCommand struct {
    light *Light
}

func (c *LightOnCommand) Execute() { c.light.On() }

type LightOffCommand struct {
    light *Light
}

func (c *LightOffCommand) Execute() { c.light.Off() }

type Remote struct {
    command Command
}

func (r *Remote) Press() { r.command.Execute() }

func main() {
    light := &Light{}
    remote := &Remote{&LightOnCommand{light}}
    remote.Press()
}
```

----------

## State
Allows object to alter behavior when state changes
```go
package main

import "fmt"

type State interface {
    Handle(c *Context)
}

type Context struct {
    state State
}

func (c *Context) SetState(s State) { c.state = s }
func (c *Context) Request()         { c.state.Handle(c) }

type IdleState struct{}

func (s *IdleState) Handle(c *Context) {
    fmt.Println("Idle -> Processing")
    c.SetState(&ProcessingState{})
}

type ProcessingState struct{}

func (s *ProcessingState) Handle(c *Context) {
    fmt.Println("Processing -> Done")
    c.SetState(&IdleState{})
}

func main() {
    ctx := &Context{&IdleState{}}
    ctx.Request()
    ctx.Request()
}
```

----------

## Chain of Responsibility
Passes request along chain of handlers
```go
package main

import "fmt"

type Handler interface {
    SetNext(Handler)
    Handle(request int)
}

type BaseHandler struct {
    next Handler
}

func (b *BaseHandler) SetNext(h Handler) { b.next = h }

type LowHandler struct{ BaseHandler }

func (h *LowHandler) Handle(r int) {
    if r < 10 {
        fmt.Printf("Low handled: %d\n", r)
    } else if h.next != nil {
        h.next.Handle(r)
    }
}

type HighHandler struct{ BaseHandler }

func (h *HighHandler) Handle(r int) {
    fmt.Printf("High handled: %d\n", r)
}

func main() {
    low := &LowHandler{}
    high := &HighHandler{}
    low.SetNext(high)
    low.Handle(5)
    low.Handle(50)
}
```

----------

## Template Method
Defines skeleton of algorithm in base, letting subclasses override steps
```go
package main

import "fmt"

type Builder interface {
    Build()
    Test()
    Deploy()
}

type Template struct {
    builder Builder
}

func (t *Template) Execute() {
    t.builder.Build()
    t.builder.Test()
    t.builder.Deploy()
}

type GoBuilder struct{}

func (g *GoBuilder) Build()  { fmt.Println("go build") }
func (g *GoBuilder) Test()   { fmt.Println("go test") }
func (g *GoBuilder) Deploy() { fmt.Println("Deploying Go app") }

func main() {
    tmpl := &Template{&GoBuilder{}}
    tmpl.Execute()
}
```

----------

## Iterator
Provides way to access elements sequentially
```go
package main

import "fmt"

type Iterator interface {
    HasNext() bool
    Next() interface{}
}

type Collection struct {
    items []string
}

type CollectionIterator struct {
    collection *Collection
    index      int
}

func (c *Collection) CreateIterator() Iterator {
    return &CollectionIterator{collection: c, index: 0}
}

func (i *CollectionIterator) HasNext() bool {
    return i.index < len(i.collection.items)
}

func (i *CollectionIterator) Next() interface{} {
    item := i.collection.items[i.index]
    i.index++
    return item
}

func main() {
    col := &Collection{items: []string{"A", "B", "C"}}
    iter := col.CreateIterator()
    for iter.HasNext() {
        fmt.Println(iter.Next())
    }
}
```

----------

## Mediator
Centralizes complex communications between objects
```go
package main

import "fmt"

type Mediator interface {
    Notify(sender string, event string)
}

type ChatRoom struct {
    users map[string]*User
}

func (c *ChatRoom) Register(u *User) {
    if c.users == nil {
        c.users = make(map[string]*User)
    }
    c.users[u.name] = u
    u.room = c
}

func (c *ChatRoom) Notify(sender, msg string) {
    for name, user := range c.users {
        if name != sender {
            user.Receive(sender, msg)
        }
    }
}

type User struct {
    name string
    room *ChatRoom
}

func (u *User) Send(msg string)              { u.room.Notify(u.name, msg) }
func (u *User) Receive(from, msg string) { fmt.Printf("%s received from %s: %s\n", u.name, from, msg) }

func main() {
    room := &ChatRoom{}
    room.Register(&User{name: "Alice"})
    room.Register(&User{name: "Bob"})
    room.users["Alice"].Send("Hello!")
}
```

----------

## Memento
Captures and restores object's internal state
```go
package main

import "fmt"

type Memento struct {
    state string
}

type Editor struct {
    content string
}

func (e *Editor) SetContent(c string) { e.content = c }
func (e *Editor) GetContent() string  { return e.content }
func (e *Editor) Save() *Memento      { return &Memento{state: e.content} }
func (e *Editor) Restore(m *Memento)  { e.content = m.state }

type History struct {
    mementos []*Memento
}

func (h *History) Push(m *Memento) { h.mementos = append(h.mementos, m) }
func (h *History) Pop() *Memento {
    if len(h.mementos) == 0 {
        return nil
    }
    m := h.mementos[len(h.mementos)-1]
    h.mementos = h.mementos[:len(h.mementos)-1]
    return m
}

func main() {
    editor := &Editor{}
    history := &History{}
    
    editor.SetContent("v1")
    history.Push(editor.Save())
    editor.SetContent("v2")
    editor.Restore(history.Pop())
    fmt.Println(editor.GetContent())
}
```

----------

## Visitor
Separates algorithm from object structure
```go
package main

import "fmt"

type Visitor interface {
    VisitCircle(*Circle)
    VisitRectangle(*Rectangle)
}

type Shape interface {
    Accept(Visitor)
}

type Circle struct{ Radius float64 }

func (c *Circle) Accept(v Visitor) { v.VisitCircle(c) }

type Rectangle struct{ Width, Height float64 }

func (r *Rectangle) Accept(v Visitor) { v.VisitRectangle(r) }

type AreaVisitor struct{ Total float64 }

func (a *AreaVisitor) VisitCircle(c *Circle) {
    a.Total += 3.14 * c.Radius * c.Radius
}

func (a *AreaVisitor) VisitRectangle(r *Rectangle) {
    a.Total += r.Width * r.Height
}

func main() {
    shapes := []Shape{&Circle{5}, &Rectangle{4, 3}}
    visitor := &AreaVisitor{}
    for _, s := range shapes {
        s.Accept(visitor)
    }
    fmt.Printf("Total area: %.2f\n", visitor.Total)
}
```

----------

## Flyweight
Shares common state between multiple objects
```go
package main

import "fmt"

type TreeType struct {
    name    string
    color   string
    texture string
}

type TreeFactory struct {
    types map[string]*TreeType
}

func (f *TreeFactory) GetType(name, color, texture string) *TreeType {
    key := name + color + texture
    if f.types == nil {
        f.types = make(map[string]*TreeType)
    }
    if _, ok := f.types[key]; !ok {
        f.types[key] = &TreeType{name, color, texture}
    }
    return f.types[key]
}

type Tree struct {
    x, y     int
    treeType *TreeType
}

func main() {
    factory := &TreeFactory{}
    t1 := &Tree{1, 2, factory.GetType("Oak", "Green", "Rough")}
    t2 := &Tree{3, 4, factory.GetType("Oak", "Green", "Rough")}
    fmt.Printf("Same type: %t\n", t1.treeType == t2.treeType)
}
```

----------

## Null Object
Provides default behavior instead of null checks
```go
package main

import "fmt"

type Logger interface {
    Log(msg string)
}

type ConsoleLogger struct{}

func (c *ConsoleLogger) Log(msg string) {
    fmt.Println("LOG:", msg)
}

type NullLogger struct{}

func (n *NullLogger) Log(msg string) {}

type App struct {
    logger Logger
}

func NewApp(debug bool) *App {
    if debug {
        return &App{logger: &ConsoleLogger{}}
    }
    return &App{logger: &NullLogger{}}
}

func (a *App) Run() {
    a.logger.Log("App started")
    fmt.Println("Running...")
}

func main() {
    app := NewApp(true)
    app.Run()
}
```

----------

## Object Pool
Reuses expensive objects instead of creating new ones
```go
package main

import (
    "fmt"
    "sync"
)

type Connection struct {
    ID int
}

type Pool struct {
    pool  chan *Connection
    mutex sync.Mutex
    count int
}

func NewPool(size int) *Pool {
    return &Pool{pool: make(chan *Connection, size)}
}

func (p *Pool) Get() *Connection {
    select {
    case conn := <-p.pool:
        return conn
    default:
        p.mutex.Lock()
        p.count++
        p.mutex.Unlock()
        return &Connection{ID: p.count}
    }
}

func (p *Pool) Put(conn *Connection) {
    select {
    case p.pool <- conn:
    default:
    }
}

func main() {
    pool := NewPool(2)
    c1 := pool.Get()
    fmt.Printf("Got connection %d\n", c1.ID)
    pool.Put(c1)
    c2 := pool.Get()
    fmt.Printf("Reused connection %d\n", c2.ID)
}
```

----------

## Registry
Global access point to store and retrieve objects
```go
package main

import (
    "fmt"
    "sync"
)

type Service interface {
    Name() string
}

type Registry struct {
    services map[string]Service
    mutex    sync.RWMutex
}

var registry = &Registry{services: make(map[string]Service)}

func (r *Registry) Register(s Service) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    r.services[s.Name()] = s
}

func (r *Registry) Get(name string) Service {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    return r.services[name]
}

type DBService struct{}

func (d *DBService) Name() string { return "database" }

func main() {
    registry.Register(&DBService{})
    svc := registry.Get("database")
    fmt.Printf("Found: %s\n", svc.Name())
}
```

----------

## Dependency Injection
Injects dependencies rather than creating them internally
```go
package main

import "fmt"

type Database interface {
    Query(q string) string
}

type MySQL struct{}

func (m *MySQL) Query(q string) string {
    return "MySQL: " + q
}

type PostgreSQL struct{}

func (p *PostgreSQL) Query(q string) string {
    return "PostgreSQL: " + q
}

type UserService struct {
    db Database
}

func NewUserService(db Database) *UserService {
    return &UserService{db: db}
}

func (u *UserService) GetUser(id int) string {
    return u.db.Query(fmt.Sprintf("SELECT * FROM users WHERE id=%d", id))
}

func main() {
    mysql := &MySQL{}
    service := NewUserService(mysql)
    fmt.Println(service.GetUser(1))
}
```

----------

## Semaphore
Limits concurrent access to a resource
```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type Semaphore struct {
    ch chan struct{}
}

func NewSemaphore(max int) *Semaphore {
    return &Semaphore{ch: make(chan struct{}, max)}
}

func (s *Semaphore) Acquire() { s.ch <- struct{}{} }
func (s *Semaphore) Release() { <-s.ch }

func main() {
    sem := NewSemaphore(2)
    var wg sync.WaitGroup
    
    for i := 1; i <= 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            sem.Acquire()
            fmt.Printf("Worker %d running\n", id)
            time.Sleep(100 * time.Millisecond)
            sem.Release()
        }(i)
    }
    wg.Wait()
}
```

----------

## Circuit Breaker
Prevents cascading failures in distributed systems
```go
package main

import (
    "errors"
    "fmt"
    "sync"
    "time"
)

type CircuitBreaker struct {
    failures  int
    threshold int
    timeout   time.Duration
    lastFail  time.Time
    mutex     sync.Mutex
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mutex.Lock()
    if cb.failures >= cb.threshold {
        if time.Since(cb.lastFail) < cb.timeout {
            cb.mutex.Unlock()
            return errors.New("circuit open")
        }
        cb.failures = 0
    }
    cb.mutex.Unlock()
    
    if err := fn(); err != nil {
        cb.mutex.Lock()
        cb.failures++
        cb.lastFail = time.Now()
        cb.mutex.Unlock()
        return err
    }
    return nil
}

func main() {
    cb := &CircuitBreaker{threshold: 3, timeout: 5 * time.Second}
    err := cb.Call(func() error { return errors.New("fail") })
    fmt.Printf("Error: %v\n", err)
}
```

----------

## Retry
Retries failed operations with configurable attempts
```go
package main

import (
    "errors"
    "fmt"
    "time"
)

type RetryConfig struct {
    Attempts int
    Delay    time.Duration
}

func Retry(cfg RetryConfig, fn func() error) error {
    var err error
    for i := 0; i < cfg.Attempts; i++ {
        if err = fn(); err == nil {
            return nil
        }
        fmt.Printf("Attempt %d failed: %v\n", i+1, err)
        if i < cfg.Attempts-1 {
            time.Sleep(cfg.Delay)
        }
    }
    return fmt.Errorf("failed after %d attempts: %w", cfg.Attempts, err)
}

func main() {
    attempt := 0
    err := Retry(RetryConfig{Attempts: 3, Delay: 100 * time.Millisecond}, func() error {
        attempt++
        if attempt < 3 {
            return errors.New("temporary error")
        }
        return nil
    })
    fmt.Printf("Final result: %v\n", err)
}
```

----------

## Fan-Out Fan-In
Distributes work and collects results concurrently
```go
package main

import (
    "fmt"
    "sync"
)

func fanOut(input <-chan int, workers int) []<-chan int {
    channels := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        channels[i] = worker(input)
    }
    return channels
}

func worker(input <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range input {
            out <- n * n
        }
    }()
    return out
}

func fanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for n := range c { out <- n }
        }(ch)
    }
    go func() { wg.Wait(); close(out) }()
    return out
}

func main() {
    input := make(chan int, 5)
    for i := 1; i <= 5; i++ { input <- i }
    close(input)
    
    for result := range fanIn(fanOut(input, 2)...) {
        fmt.Println(result)
    }
}
```

----------

## Pipeline
Chains processing stages with channels
```go
package main

import "fmt"

func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

func filter(in <-chan int, predicate func(int) bool) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            if predicate(n) {
                out <- n
            }
        }
    }()
    return out
}

func main() {
    nums := generate(1, 2, 3, 4, 5)
    squared := square(nums)
    filtered := filter(squared, func(n int) bool { return n > 10 })
    
    for n := range filtered {
        fmt.Println(n)
    }
}
```

----------

## Pub/Sub
Asynchronous message distribution to subscribers
```go
package main

import (
    "fmt"
    "sync"
)

type PubSub struct {
    subs  map[string][]chan string
    mutex sync.RWMutex
}

func NewPubSub() *PubSub {
    return &PubSub{subs: make(map[string][]chan string)}
}

func (ps *PubSub) Subscribe(topic string) <-chan string {
    ps.mutex.Lock()
    defer ps.mutex.Unlock()
    ch := make(chan string, 1)
    ps.subs[topic] = append(ps.subs[topic], ch)
    return ch
}

func (ps *PubSub) Publish(topic, msg string) {
    ps.mutex.RLock()
    defer ps.mutex.RUnlock()
    for _, ch := range ps.subs[topic] {
        ch <- msg
    }
}

func main() {
    ps := NewPubSub()
    sub := ps.Subscribe("news")
    go func() { ps.Publish("news", "Hello!") }()
    fmt.Println(<-sub)
}
```
