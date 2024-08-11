# cargo

My attempt to implement [the DDD Sample App](https://github.com/citerus/dddsample-core) in idiomatic Golang, 
so the project could be used to showcase the idiomatic project structuring, namings, etc.

At the same time, it's not the goal to blindly follow the DDD pracrices nor clean architecture. The goal is to identify the Go's way.

## Requirements

- As simple and idiomatic as possible.
- The codebase should be testable (a lot of code should be tested via simple unit tests).
- The layout should be suitable for middle to large services (like 100 of API endpoints, think about a startup with a single monolitic backend server).

## References

- [marcusolsson/goddd](https://github.com/marcusolsson/goddd)
- [ThreeDotsLabs/wild-workouts-go-ddd-example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
- [eyazici90/go-ddd](https://github.com/eyazici90/go-ddd/tree/master)
