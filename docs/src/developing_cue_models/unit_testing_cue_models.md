# Unit Testing CUE Models

Unit testing CUE models is a crucial part of the development process, allowing you to validate the logic and behavior of your models in isolation, before publishing them for use. This can help catch errors early, ensure your models behave as expected, and provide confidence in their correctness.

Cuestomize, being built as a Go library, allows you to write unit tests for your CUE models using Go's testing framework.

The following code blocks show an example of how to write unit tests for a CUE model using the Cuestomize library.

```go
{{#include ../../../examples/validation/model_test.go:32:89}}
```

The test files can live in the same repository as your CUE model, and you can use a local path provider to point the Cuestomize function to the local directory where your CUE model is located. This allows you to test your CUE model without needing to publish it to an OCI registry, enabling a fast development and testing cycle.
