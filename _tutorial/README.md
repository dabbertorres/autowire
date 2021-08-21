# Autowire Tutorial

Let's learn to use Autowire by example. The [Autowire guide][guide] provides thorough
documentation of the tool's usage. For readers eager to see Autowire applied to a
larger server, the [guestbook sample in Go Cloud][guestbook] uses Autowire to
initialize its components. Here we are going to build a small greeter program to
understand how to use Autowire. The finished product may be found in the same
directory as this README.

[guestbook]: https://github.com/google/go-cloud/tree/master/samples/guestbook
[guide]:     https://github.com/google/autowire/blob/master/docs/guide.md

## A First Pass of Building the Greeter Program

Let's create a small program that simulates an event with a greeter greeting
guests with a particular message.

To start, we will create three types: 1) a message for a greeter, 2) a greeter
who conveys that message, and 3) an event that starts with the greeter greeting
guests. In this design, we have three `struct` types:

``` go
type Message string

type Greeter struct {
    // ... TBD
}

type Event struct {
    // ... TBD
}
```

The `Message` type just wraps a string. For now, we will create a simple
initializer that always returns a hard-coded message:

``` go
func NewMessage() Message {
    return Message("Hi there!")
}
```

Our `Greeter` will need reference to the `Message`. So let's create an
initializer for our `Greeter` as well.

``` go
func NewGreeter(m Message) Greeter {
    return Greeter{Message: m}
}

type Greeter struct {
    Message Message // <- adding a Message field
}
```

In the initializer we assign a `Message` field to `Greeter`. Now, we can use the
`Message` when we create a `Greet` method on `Greeter`:

``` go
func (g Greeter) Greet() Message {
    return g.Message
}
```

Next, we need our `Event` to have a `Greeter`, so we will create an initializer
for it as well.

``` go
func NewEvent(g Greeter) Event {
    return Event{Greeter: g}
}

type Event struct {
    Greeter Greeter // <- adding a Greeter field
}
```

Then we add a method to start the `Event`:

``` go
func (e Event) Start() {
    msg := e.Greeter.Greet()
    fmt.Println(msg)
}
```

The `Start` method holds the core of our small application: it tells the
greeter to issue a greeting and then prints that message to the screen.

Now that we have all the components of our application ready, let's see what it
takes to initialize all the components without using Autowire. Our main function
would look like this:

``` go
func main() {
    message := NewMessage()
    greeter := NewGreeter(message)
    event := NewEvent(greeter)

    event.Start()
}
```

First we create a message, then we create a greeter with that message, and
finally we create an event with that greeter. With all the initialization done,
we're ready to start our event.

We are using the [dependency injection][di] design principle. In practice, that
means we pass in whatever each component needs. This style of design lends
itself to writing easily tested code and makes it easy to swap out one
dependency with another.

[di]: https://stackoverflow.com/questions/130794/what-is-dependency-injection

## Using Autowire to Generate Code

One downside to dependency injection is the need for so many initialization
steps. Let's see how we can use Autowire to make the process of initializing our
components smoother.

Let's start by changing our `main` function to look like this:

``` go
func main() {
    e := InitializeEvent()

    e.Start()
}
```

Next, in a separate file called `autowire.go` we will define `InitializeEvent`.
This is where things get interesting:

``` go
// autowire.go

func InitializeEvent() Event {
    autowire.Build(NewEvent, NewGreeter, NewMessage)
    return Event{}
}
```

Rather than go through the trouble of initializing each component in turn and
passing it into the next one, we instead have a single call to `autowire.Build`
passing in the initializers we want to use. In Autowire, initializers are known as
"providers," functions which provide a particular type. We add a zero value for
`Event` as a return value to satisfy the compiler. Note that even if we add
values to `Event`, Autowire will ignore them. In fact, the injector's purpose is to
provide information about which providers to use to construct an `Event` and so
we will exclude it from our final binary with a build constraint at the top of
the file:

``` go
//+build wireinject
//go:build wireinject

```

Note, a [build constraint][constraint] requires a blank, trailing line.

In Autowire parlance, `InitializeEvent` is an "injector." Now that we have our
injector complete, we are ready to use the `autowire` command line tool.

Install the tool with:

``` shell
go get github.com/dabbertorres/autowire/cmd/autowire
```

Then in the same directory with the above code, simply run `autowire`. Autowire will
find the `InitializeEvent` injector and generate a function whose body is
filled out with all the necessary initialization steps. The result will be
written to a file named `autowire_gen.go`.

Let's take a look at what Autowire did for us:

``` go
// autowire_gen.go

func InitializeEvent() Event {
    message := NewMessage()
    greeter := NewGreeter(message)
    event := NewEvent(greeter)
    return event
}
```

It looks just like what we wrote above! Now this is a simple example with just
three components, so writing the initializer by hand isn't too painful. Imagine
how useful Autowire is for components that are much more complex. When working with
Autowire, we will commit both `autowire.go` and `autowire_gen.go` to source control.

[constraint]: https://godoc.org/go/build#hdr-Build_Constraints

## Making Changes with Autowire

To show a small part of how Autowire handles more complex setups, let's refactor
our initializer for `Event` to return an error and see what happens.

``` go
func NewEvent(g Greeter) (Event, error) {
    if g.Grumpy {
        return Event{}, errors.New("could not create event: event greeter is grumpy")
    }
    return Event{Greeter: g}, nil
}
```

We'll say that sometimes a `Greeter` might be grumpy and so we cannot create
an `Event`. The `NewGreeter` initializer now looks like this:

``` go
func NewGreeter(m Message) Greeter {
    var grumpy bool
    if time.Now().Unix()%2 == 0 {
        grumpy = true
    }
    return Greeter{Message: m, Grumpy: grumpy}
}
```

We have added a `Grumpy` field to `Greeter` struct and if the invocation time
of the initializer is an even number of seconds since the Unix epoch, we will
create a grumpy greeter instead of a friendly one.

The `Greet` method then becomes:

``` go
func (g Greeter) Greet() Message {
    if g.Grumpy {
        return Message("Go away!")
    }
    return g.Message
}
```

Now you see how a grumpy `Greeter` is no good for an `Event`. So `NewEvent` may
fail. Our `main` must now take into account that `InitializeEvent` may in fact
fail:

``` go
func main() {
    e, err := InitializeEvent()
    if err != nil {
        fmt.Printf("failed to create event: %s\n", err)
        os.Exit(2)
    }
    e.Start()
}
```

We also need to update `InitializeEvent` to add an `error` type to the return value:

``` go
// autowire.go

func InitializeEvent() (Event, error) {
    autowire.Build(NewEvent, NewGreeter, NewMessage)
    return Event{}, nil
}
```

With the setup complete, we are ready to invoke the `autowire` command again. Note,
that after running `autowire` once to produce a `autowire_gen.go` file, we may also use
`go generate`. Having run the command, our `autowire_gen.go` file looks like
this:

``` go
// autowire_gen.go

func InitializeEvent() (Event, error) {
    message := NewMessage()
    greeter := NewGreeter(message)
    event, err := NewEvent(greeter)
    if err != nil {
        return Event{}, err
    }
    return event, nil
}
```

Autowire has detected that the `NewEvent` provider may fail and has done the right
thing inside the generated code: it checks the error and returns early if one
is present.

## Changing the Injector Signature

As another improvement, let's look at how Autowire generates code based on the
signature of the injector. Presently, we have hard-coded the message inside
`NewMessage`. In practice, it's much nicer to allow callers to change that
message however they see fit. So let's change `InitializeEvent` to look like
this:

``` go
func InitializeEvent(phrase string) (Event, error) {
    autowire.Build(NewEvent, NewGreeter, NewMessage)
    return Event{}, nil
}
```

Now `InitializeEvent` allows callers to pass in the `phrase` for a `Greeter` to
use. We also add a `phrase` argument to `NewMessage`:

``` go
func NewMessage(phrase string) Message {
    return Message(phrase)
}
```

After we run `autowire` again, we will see that the tool has generated an
initializer which passes the `phrase` value as a `Message` into `Greeter`.
Neat!

``` go
// autowire_gen.go

func InitializeEvent(phrase string) (Event, error) {
    message := NewMessage(phrase)
    greeter := NewGreeter(message)
    event, err := NewEvent(greeter)
    if err != nil {
        return Event{}, err
    }
    return event, nil
}
```

Autowire inspects the arguments to the injector, sees that we added a string to the
list of arguments (e.g., `phrase`), and likewise sees that among all the
providers, `NewMessage` takes a string, and so it passes `phrase` into
`NewMessage`.

## Catching Mistakes with Helpful Errors

Let's also look at what happens when Autowire detects mistakes in our code and see
how Autowire's error messages help us correct any problems.

For example, when writing our injector `InitializeEvent`, let's say we forget
to add a provider for `Greeter`. Let's see what happens:

``` go
func InitializeEvent(phrase string) (Event, error) {
    autowire.Build(NewEvent, NewMessage) // woops! We forgot to add a provider for Greeter
    return Event{}, nil
}
```

Running `autowire`, we see the following:

``` shell
# wrapping the error across lines for readability
$GOPATH/src/github.com/google/autowire/_tutorial/autowire.go:24:1:
inject InitializeEvent: no provider found for github.com/google/autowire/_tutorial.Greeter
(required by provider of github.com/google/autowire/_tutorial.Event)
autowire: generate failed
```

Autowire is telling us some useful information: it cannot find a provider for
`Greeter`. Note that the error message prints out the full path to the
`Greeter` type. It's also telling us the line number and injector name where
the problem occurred: line 24 inside `InitializeEvent`. In addition, the error
message tells us which provider needs a `Greeter`. It's the `Event` type. Once
we pass in a provider of `Greeter`, the problem will be solved.

Alternatively, what happens if we provide one too many providers to `autowire.Build`?

``` go
func NewEventNumber() int  {
    return 1
}

func InitializeEvent(phrase string) (Event, error) {
     // woops! NewEventNumber is unused.
    autowire.Build(NewEvent, NewGreeter, NewMessage, NewEventNumber)
    return Event{}, nil
}
```

Autowire helpfully tells us that we have an unused provider:

``` shell
$GOPATH/src/github.com/google/autowire/_tutorial/autowire.go:24:1:
inject InitializeEvent: unused provider "NewEventNumber"
autowire: generate failed
```

Deleting the unused provider from the call to `autowire.Build` resolves the error.

## Conclusion

Let's summarize what we have done here. First, we wrote a number of components
with corresponding initializers, or providers. Next, we created an injector
function, specifying which arguments it receives and which types it returns.
Then, we filled in the injector function with a call to `autowire.Build` supplying
all necessary providers. Finally, we ran the `autowire` command to generate code
that wires up all the different initializers. When we added an argument to the
injector and an error return value, running `autowire` again made all the necessary
updates to our generated code.

The example here is small, but it demonstrates some of the power of Autowire, and
how it takes much of the pain out of initializing code using dependency
injection. Furthermore, using Autowire produced code that looks much like what we
would otherwise write. There are no bespoke types that commit a user to Autowire.
Instead it's just generated code. We may do with it what we will. Finally,
another point worth considering is how easy it is to add new dependencies to
our component initialization. As long as we tell Autowire how to provide (i.e.,
initialize) a component, we may add that component anywhere in the dependency
graph and Autowire will handle the rest.

In closing, it is worth mentioning that Autowire supports a number of additional
features not discussed here. Providers may be grouped in [provider sets][sets].
There is support for [binding interfaces][interfaces], [binding
values][values], as well as support for [cleanup functions][cleanup]. See the
[Advanced Features][advanced] section for more.

[advanced]:   https://github.com/google/autowire/blob/master/docs/guide.md#advanced-features
[cleanup]:    https://github.com/google/autowire/blob/master/docs/guide.md#cleanup-functions
[interfaces]: https://github.com/google/autowire/blob/master/docs/guide.md#binding-interfaces
[sets]:       https://github.com/google/autowire/blob/master/docs/guide.md#defining-providers
[values]:     https://github.com/google/autowire/blob/master/docs/guide.md#binding-values
