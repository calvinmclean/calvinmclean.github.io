REST, or Representational State Transfer, is a concept that guides the implementation of an application's interface. It is typically used in HTTP communication with JSON data to define create, read, update and delete (CRUD) operations on resources. However, by definition, REST APIs can exist without JSON and even without HTTP. Incorrectly claiming that _any_ HTTP JSON API as REST is harmful to the standard since the assumptions about the API's implementation might not be true. In addition to this, non-standard APIs require extensive documentation in order to be usable. If we correctly implement REST and use accurate terminology when discussing APIs, we can build applications that are much easier to use and maintain.

## Basics of REST

Roy Fielding defined the term REST in his [1999-2000 PhD dissertation](https://ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm). Fundamentally, REST requires:
  - Client-server architecture
  - Statelessness between requests
  - Uniform interface
  - Identifiable resources
  - Hypermedia representations

The first few requirements are often achieved using HTTP methods (`GET`, `POST`, `PUT`, `DELETE`) and URIs (Uniform Resource Identifiers) as the uniform interface and identification of resources. This clear and consistent interaction builds confidence in which resource is affected by the request and what that effect is. The hypermedia can be JSON, HTML, or anything else that can provide a consistent representation of the application’s resources. 

As we dig deeper into REST, the examples will describe an HTTP and JSON API for the sake of familiarity and wide-appeal, but keep in mind that these are not defining characteristics of REST.


## REST API Maturity

To better commmunicate the level of RESTfulness that an HTTP API achieves, the [Richardson maturity model](https://martinfowler.com/articles/richardsonMaturityModel.html) describes four levels:

- _Level 0_: Use a URI and POST for client-server interaction
- _Level 1_: Use separate URIs for each resource
- _Level 2_: Introduce more HTTP methods for different operations on resources. Follow established standards for the behavior of each method
- _Level 3_: Use hypermedia

Notice that we are now including HTTP-specifics like methods and URIs into the requirements since this maturity model is specific for REST APIs on the web, not the more general definition of REST.

Level 2 is the most impactful level since it introduces a lot of functionality and consistency to the API. It also dictates the use of standard HTTP response codes for communicating the success or failure of requests, such as `200 OK` and `400 Bad Request`. Once your API reaches this level, it demonstrates a fully-functioning application that allows for all necessary interactions between the client and the server. In most cases, this level of API maturity is sufficient, but what does it look like to take it one step further?

The last, and often overlooked, piece is hypermedia linking, or Hypertext As The Engine of Application State (HATEOAS). Despite the horrendous acronym, this level introduces sophistacation and friendliness to an already-useful API. By adding links to the responses, the server creates a self-documenting API where the client is free to navigate the entire interface using the base path and reasonable assumptions about HTTP methods.

For example, if we are navigating an API for blog posts, we can start with a request to `GET /posts`. If the API is RESTful, it will respond with a list of Post resources containing details and a link to `/posts/{id}`. The server will only provide the client with Posts that it has permission to access, so the client can follow the link to `GET /posts/1`, which will respond with details of the Post and, ideally, links to more related resources such as Likes and Comments:
```json
{
	"id": 1,
	"title": "Post Title",
	...,
	"links": {
		"comments": "/posts/1/comments",
		"likes": "/posts/1/likes"
	}
}
```

As you can see from this example, the basic knowledge of the `/posts` endpoint and the `GET` method is all the client needs to fully navigate the Posts API. A level-2 API would not enable the client to discover the Comments or Likes related to a Post. If you look beyond the unpronounceable acronym, HATEOAS is an essential feature of a robust REST API.

This example uses the more-familiar JSON API, but the same thing can be achieved using HTML as the hypermedia. In this case, the hypermedia would reveal the child resources by actual clickable links or UI elements instead of just the URIs. Implementations of REST clearly go beyond structural JSON representations of objects and instead show us more generally how to interact with an application.


## Self Documenting

You can create a very functional API without level 3, but only the root-level resources will be available to clients without some sort of external documentation or brute-force search. In most cases, this is achieved using an OpenAPI specification.

OpenAPI specifications are a generic and expressive way to document APIs, but they are tedious to write and keep in sync with actual API behavior. The extra step of maintaining API specifications on top of writing code and managing other responsibilities allows the implementation and specification to drift apart. It is worth considering if we can skip this step altogether when our APIs are intuitive and standardized.

Query parameters and request headers are overlooked by self-documenting REST principles. OpenAPI Specifications come out ahead in this aspect, but self-documenting APIs can still be very usable if they adhere to a few assumptions:
  - A resource's root endpoint (like `GET /posts`) can filter/search by any field of a Post, like title or author
  - There might be more fields that aid in querying like `order` and `limit`
  - User-friendly errors will make troubleshooting easy

Before you start deleting all of your API specifications, remember every scenario has unique considerations. We can't realistically expect to learn every API by trial-and-error, so external documentation is still valuable. Thoughtful implementation of REST builds a solid foundation for a usable and well-documented application. Then, API specifications are a valuable add-on to provide more clarity about the API. However, the foundations of REST come first and a specification is not an excuse for eschewing standards. Ultimately, you should find a balance that works for you, your peers, and users.


## babyapi

Now this is the part where I tell you about a library I created, [`babyapi`](https://github.com/calvinmclean/babyapi), and how it makes all of this tedious REST stuff easy!

I created babyapi with the goal of providing the [simplest path to a full REST API](https://dev.to/calvinmclean/the-easiest-way-to-create-a-rest-api-with-go-20bo). It started as a humble package that used generics to reduce code duplication in my [`automated-garden`](https://github.com/calvinmclean/automated-garden) API. After breaking it out into its own repository, I added the optional `HATEOAS` extension which sets up automatic hypermedia linking. Now you can easily create a REST API which achieves all levels in the Richardson maturity model using `babyapi`.

[Check it out](https://github.com/calvinmclean/babyapi/tree/main/examples/nested) and let me know what you think!


## Conclusion

REST is a useful standard for creating intuitive application interfaces, especially in the case of HTTP web applications. In modern web development, the understood definition of REST has drifted to just mean any JSON HTTP interface that loosely interacts with resources. In reality, these are not adhering to the strict definitions of REST and claiming RESTfulness is harmful to usability and maintainability of the APIs. While API specifications are a powerful tool for describing complex APIs, I hope you will first consider designing a self-documenting REST API. This will enable your users to quickly get started with your API before getting bogged down by specifications. Of course, specifications still have an appropriate role in many complex APIs, but we should not be relying on them to make an API usable in the first place.
