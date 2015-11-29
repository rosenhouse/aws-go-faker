# awsfaker
[![Build Status](https://api.travis-ci.org/rosenhouse/awsfaker.png?branch=master)](http://travis-ci.org/rosenhouse/awsfaker)
[![GoDoc](https://godoc.org/github.com/rosenhouse/awsfaker?status.svg)](https://godoc.org/github.com/rosenhouse/awsfaker)

A Go library for faking AWS over the network

### Quick start
Check out the [example](example_test.go)

### Context
Integration testing of applications that use AWS can be difficult.  A test suite that interacts with a live AWS account will provide good test coverage, but may be slow and expensive.

An alternative is to create a test double or "fake" of the AWS APIs that your application uses.  The fake boots an HTTP server that stands in for the real AWS endpoints, recording requests and providing arbitrary responses.

This package provides a generic HTTP handler that can form the front-end of a test double (mock, fake or stub) for an AWS API.

### Usage
1. Build a "backend" that implements the subset of the AWS API used by your code.
  
  Each API call should be implemented as a backend method with a signature like
  ```go
  func (b *MyBackend) SomeAction(input *service.SomeActionInput) (*service.SomeActionOutput, error)
  ```
  
2. During test setup, initialize an HTTP test server that wraps your backend
  ```go
  myBackend := &MyBackend{ ... }
  fakeServer := httptest.NewServer(awsfaker.New(myBackend))
  ```
  
3. In your tests, configure the client that you are testing to use your fake server instead of the real AWS
  ```go
  app := myapp.App{ AWSOverride: fakeServer.URL }
  app.Run()
  ```

The method signatures expected on the backends match the patterns of [aws-sdk-go](https://github.com/aws/aws-sdk-go).  For example, a complete implementation of AWS CloudFormation would match the [CloudFormationAPI interface](https://github.com/aws/aws-sdk-go/blob/master/service/cloudformation/cloudformationiface/interface.go)

But your backend need only implement those methods used by your code under test.
