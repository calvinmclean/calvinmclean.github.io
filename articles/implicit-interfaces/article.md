## Introduction

DRY, or Don't Repeat Yourself, is one of the most well-known software development principles. Many engineers will go to great lengths to achieve this laudable goal, so a successful programming language must be designed with this in mind. Interfaces are just one of the tools that programming languages implement to enable code reuse.

Although there are countless differences in the way programming languages implement interfaces, they all define a contract for interaction between software components. Differentiating itself from the typical object-oriented implementation, the Go programming language implements interfaces in a way that aims to reduce complexity, encourage composition, and enable flexible use of the types.

## Implicit Interfaces

The most impactful difference between Go’s interfaces and those from object-oriented languages is that Go’s interfaces work by implicit implementation. Rather than requiring a type to explicitly declare that it implements an interface, Go determines that a type satisfies an interface as long as it implements all of the required method signatures. The different types and the interface they implement are only coupled by the code using them, rather than always being linked together.

This difference offers a few key benefits:
- Reduce complexity and coupling by limiting hierarchical inheritance of interfaces
- Simplify types that implement multiple interfaces
- Enables you to add interfaces later on when they become necessary instead of designing entire applications around them
- Interfaces can be created for types in different packages

This last benefit is the most interesting and unique. The ability to define an interface based on external types leads to a few more specific advantages:
- Client code can define how it will use types rather than relying on library code to tell it how types must be used
- Create polymorphic relationships with types that you do not own or cannot change, which allows you to use them more flexibly
- Solve import cycle issues by changing the location of interface definitions
- Create mocks of imported types to make your code more testable

## No Inheritance

By excluding inheritance, Go reduces the complexity that can occur from a deep hierarchical structure. When programs are designed around a base set of classes or interfaces, any simple changes to those base structures requires a significant refactor.

The alternative practice of composition leads to reduced complexity and more readable code. Composition relies on splitting up functionality among different types, and using them together, instead of re-defining the functionality of types through inheritance. Now you are able to re-use these individual components elsewhere, add more functionality with new components, and easily refactor or remove exiting ones.

Instead of being concerned about what type something _is_, your code just needs to known about what that type _can do_, and luckily the interface informs it.

## Use Case: Polymorphism

Polymorphism is perhaps the entire reason behind the existence of interfaces. This common practice is one of the most effective and easy-to-use methods of code reuse. Since an interface defines a strict contract for how types are used, these different types can be used interchangeably; this is polymorphism.

A very common and useful scenario for this is having a flexible storage backend for your program: use Postgres in production, SQLite when running locally, and mocks when testing (or skip the database mocks, but that's a topic for another day).

```go
type StorageClient interface {
    GetValue(id string) (string, error)
}

func NewStorageClient(clientType string) (StorageClient, error) {
    switch clientType {
        case "sqlite":
            return sqlite.NewClient()
        case "postgres":
            return postgres.NewClient()
        default:
            return nil, fmt.Errorf("invalid client type: %s", clientType)
    }
}
```

This implementation allows you to easily use the `StorageClient` interface throughout the program without concern for the data storage layer behind it.

## Use Case: Testing and Mocks

You can take advantage of the implicit nature of interfaces by defining an interface for the functions you use from an external library. For example, you are assigned a task to implement a function that fetches recent rain data, in inches, from a weather data API. The made-up weather provider publishes Go package called `weather`, which provides a `Client` struct with various weather-related methods returning data in metric units:

```go
func GetRainInches(since time.Duration, client weather.Client) (float32, error) {
    rainMM, err := client.GetRain(since)
    if err != nil {
        return 0, fmt.Errorf("error getting data from API: %w", err)
    }

    return rainMM / 2.54, nil
} 
```

How will you unit test this code? Since Go has implicit interfaces, you can create your own interface that just defines the methods that you need from the library. Now, if your function expects this interface instead, you can create your own mocks. Since you currently just need the `GetRain` method, this is really simple:
```go
type WeatherClient interface {
    GetRain(time.Duration) (float32, error)
}

func GetRainInches(since time.Duration, client WeatherClient) (float32, error) {
    rainMM, err := client.GetRain(since)
    if err != nil {
        return 0, fmt.Errorf("error getting data from API: %w", err)
    }

    return rainMM / 2.54, nil
}
```
Then, your test file might contain a new struct that also implements the interface:
```go
type MockWeatherClient struct {
    expectedErr error
    expectedMM  float32
}

func (c MockWeatherClient) GetRain(time.Duration) (float32, error) {
    return c.expectedMM, c.expectedErr
}
```
After this simple refactor, you do not depend on external libraries to provide interfaces that make your own code testable! You have the additional side-effect of being one step closer to allowing your program to use different weather APIs generically.

# Conclusion and Warnings

While interfaces in Go were designed with considerations for simplicity and avoiding some of the common pitfalls of object-oriented patterns, there are still some things to be aware of.

It may be tempting to define interfaces for everything with the hopes that you can create a more generic and flexible program. Remember that one of implicit interfaces is that you can easily create a new interface when you need it without making changes to any existing types, so there is no benefit to creating interfaces early in the process. Additionally, since types do not explicitly declare which interfaces they implement, it may be hard to tell which types are actually being used by your program when you have superfluous interfaces.

While there are always tradeoffs and no perfect solutions, I have found Go's version of interfaces to be incredibly flexible, intuitive, and useful. I have been able to create more flexible programs and improve testability all while minimizing complexity.