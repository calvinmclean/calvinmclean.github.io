## Introduction

Test-driven development is an effective method for ensuring well-tested and refactorable code. The basic idea is that you start development by writing tests. These tests clearly document expectations and create a rubric for a successful implementation. When done properly, you can clearly define the expected input/output of a function before writing any code. This has a few immediate benefits:
  - You carefully consider the interface for interacting with your code and design it to be testable
  - When you begin writing code, your flow isn't interrupted by manual testing or stepping through execution logic to predict the outcome. Instead, you just run the tests
  - Making a test pass becomes a goal that is satisfying to achieve. Breaking down the process into a series of well-defined and achieveable milestones makes the work more enjoyable
  - Avoid post-implementation laziness and over-confidence that could prevent you from testing your code

Now that you're convinced of the benefits, you can get started with test-driven development (TDD) by following these steps:
  1. Write or modify tests
  2. Check if test fails
  3. Write the minimum amount of code to make tests pass

These steps are followed in a cycle so you are always adding more tests to challenge the current implementation.

The last step, which specifies writing the minimum amount of code, is where things can get tedious if followed rigidly. It's important to understand why this rule exists before you can determine when it's appropriate to stray from it.


## Simple Example

You're tasked with implementing the function `Add(x, y int) int`. Before you jump to the implementation and just `return x + y`, write the simplest test: `1 + 1 == 2`. Then, what is the simplest implementation that would pass the test? It's just `return 2`. Now your tests pass!

At this point, you realize that you need more tests, so you pick up the pace and add a few more:
- `1 + 2 == 3`
- `100 + 5 == 105`

Now your tests fail, so you need to fix the implementation. You can't just `return 3` or `return 105` this time, so you need to find a solution that works for all tests. This leads to the implementation: `return x + y`.

While this feels overly tedious in the trivial example, strict adherence to this method caused you to write multiple tests instead of just trusting your implementation. Of course, your initial idea to `return x + y` would have worked, but the point is to re-train yourself to rely on tests rather than your own understanding of the code. In the real world, you're not the only one working on this piece of code and will inevitably forget implementation details. This process forces you to write more tests and think of more ways to break the simple implementation.

Eventually, you'll gain experience and learn to find the balance that works in the different scenarios that you encounter. You'll get back to full-speed implementation of features and find that you have fewer bugs and write more maintanable code.


## Step by step TDD for an HTTP API

Let's get into a more complicated example using TDD for an HTTP REST API. This step-by-step guide uses my Go framework, [`babyapi`](https://github.com/calvinmclean/babyapi), but the concepts can be applied anywhere.

`babyapi` uses generics to create a full CRUD API around Go structs, making it super easy to create a full REST API and client CLI. In addition to this, the `babytest` package provides some tools for creating end-to-end API tables tests. Using TDD at the API-level allows for fully testing the HTTP and storage layers of a new API or feature all at once.

Disclaimer: Since `babyapi` handles most of the implementation and also is used to generate test boilerplate, we aren't technically starting with TDD. However, we'll see how beneficial it is when adding support for `PATCH` requests to our API.

1. Create a new Go project
{% embed https://gist.github.com/calvinmclean/be7fa26193cc67ccaaa63ef28555df7c  %}

2. Create initial `main.go` using [`babyapi`'s simple example](https://github.com/calvinmclean/babyapi/blob/main/examples/simple/main.go)
{% embed https://gist.github.com/calvinmclean/aceb82ebf1983a89fe16fb0b20260122 %}

3. Use the CLI to generate a [test boilerplate](https://gist.github.com/calvinmclean/16fcc97d8e9f2fe30b8d0f7c44243a24)
{% embed https://gist.github.com/calvinmclean/a501e975ce3cd8eb7ea6843a5ecae9a5 %}

4. Implement each test by filling in the placeholders with expected JSON
{% embed https://gist.github.com/calvinmclean/68dfca1ff7bd5460f2991c958eb0b418 %}

5. Run the tests and see that they pass!

6. Since `PUT` is idempotent, it requires all fields to be included. To avoid this, we want to add support for toggling `Completed` with `PATCH` requests. We start by adding a simple test for what we expect this feature to look like
{% embed https://gist.github.com/calvinmclean/2be84d9a018771133966d2d5277a800b %}

7. This test fails since `babyapi` doesn't support `PATCH` by default. We can fix it by implementing `Patch` for the `TODO` struct. Since we defined our feature with two tests, our simplest implementation isn't just setting `Completed = true` and we have to use the value from the request
{% embed https://gist.github.com/calvinmclean/8868cf88584dcbd874b5bcd2f68a7e78 %}

8. Now we can change the `Completed` status of a `TODO`, but we still cannot use `PATCH` to modify other fields as show by this new set of tests
{% embed https://gist.github.com/calvinmclean/21ad0cee532b4d6ddff94b9344737d98 %}

9. Update `Patch` to set the remaining fields
{% embed https://gist.github.com/calvinmclean/d0ad917cfa4862f10958fbc5ab1b37a4 %}

10. Our tests still fail since we always update the `TODO` with the request fields, even if they're empty. Fix this by updating the implementation to check for empty values
{% embed https://gist.github.com/calvinmclean/3d158228657d0090b75aa9f4167f4370 %}

11. The new `UpdateWithPatch` test passes, but our previous tests fail. Since we changed `Completed` to be `*bool`, `TODO`s created with an empty value will show as `null`
{% embed https://gist.github.com/calvinmclean/85044ad6d3f8e12bf6e42802b9d4172e %}

12. Implement `Render` for `TODO` so we can treat `nil` as `false`
{% embed https://gist.github.com/calvinmclean/17e767b09d73137b0fb6183d66239145 %}

Implementing the `PATCH` feature with test-driven development resulted in a robust set of tests and a well-implemented feature. Since we started by defining the expected input and output of a `PATCH` request in tests, it was easy to see the issues caused by not checking for empty values in the request. Also, our pre-existing tests were able to protect from breaking changes when changing the type of `Completed` to `*bool`.


## Conclusion

Test-driven development is an effective approach for creating fully tested and correct code. By starting with tests in mind, we can ensure that every piece of code is designed to be testable instead of letting tests be an afterthought.

If you're hesitant about adopting TDD, here are a few ideas to get started:
- Try it in simple scenarios where a function's input/output is clear and the implementation is not overly complicated. You can write a robust table test for the variety of input/output that could be encountered. Having a clear visual of the different scenarios can simplify implementation
- If you're fixing a new bug, you have already identified a gap in your testing. Start by writing a test that would have identified this bug in the first place. Then, make this test pass without breaking any existing tests.
- Similar to the `babyapi` example, you can use TDD for high-level API tests. Once you have a definition of the expected request/response, you can resume your usual development flow for more detail-oriented parts of the implementation

Even if TDD isn't a good fit for the way you write code, it's still a powerful tool to have in your belt. I encourage you to at least commit some time to trying it out and see how it affects your development process.
