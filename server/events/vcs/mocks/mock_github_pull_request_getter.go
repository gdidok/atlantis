// Code generated by pegomock. DO NOT EDIT.
// Source: github.com/runatlantis/atlantis/server/events/vcs (interfaces: GithubPullRequestGetter)

package mocks

import (
	github "github.com/google/go-github/v31/github"
	pegomock "github.com/petergtz/pegomock"
	models "github.com/runatlantis/atlantis/server/events/models"
	"reflect"
	"time"
)

type MockGithubPullRequestGetter struct {
	fail func(message string, callerSkip ...int)
}

func NewMockGithubPullRequestGetter(options ...pegomock.Option) *MockGithubPullRequestGetter {
	mock := &MockGithubPullRequestGetter{}
	for _, option := range options {
		option.Apply(mock)
	}
	return mock
}

func (mock *MockGithubPullRequestGetter) SetFailHandler(fh pegomock.FailHandler) { mock.fail = fh }
func (mock *MockGithubPullRequestGetter) FailHandler() pegomock.FailHandler      { return mock.fail }

func (mock *MockGithubPullRequestGetter) GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error) {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockGithubPullRequestGetter().")
	}
	params := []pegomock.Param{repo, pullNum}
	result := pegomock.GetGenericMockFrom(mock).Invoke("GetPullRequest", params, []reflect.Type{reflect.TypeOf((**github.PullRequest)(nil)).Elem(), reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 *github.PullRequest
	var ret1 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(*github.PullRequest)
		}
		if result[1] != nil {
			ret1 = result[1].(error)
		}
	}
	return ret0, ret1
}

func (mock *MockGithubPullRequestGetter) GetPullRequestFromName(repoName string, repoOwner string, pullNum int) (*github.PullRequest, error) {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockGithubPullRequestGetter().")
	}
	params := []pegomock.Param{repoName, repoOwner, pullNum}
	result := pegomock.GetGenericMockFrom(mock).Invoke("GetPullRequestFromName", params, []reflect.Type{reflect.TypeOf((**github.PullRequest)(nil)).Elem(), reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 *github.PullRequest
	var ret1 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(*github.PullRequest)
		}
		if result[1] != nil {
			ret1 = result[1].(error)
		}
	}
	return ret0, ret1
}

func (mock *MockGithubPullRequestGetter) VerifyWasCalledOnce() *VerifierMockGithubPullRequestGetter {
	return &VerifierMockGithubPullRequestGetter{
		mock:                   mock,
		invocationCountMatcher: pegomock.Times(1),
	}
}

func (mock *MockGithubPullRequestGetter) VerifyWasCalled(invocationCountMatcher pegomock.InvocationCountMatcher) *VerifierMockGithubPullRequestGetter {
	return &VerifierMockGithubPullRequestGetter{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
	}
}

func (mock *MockGithubPullRequestGetter) VerifyWasCalledInOrder(invocationCountMatcher pegomock.InvocationCountMatcher, inOrderContext *pegomock.InOrderContext) *VerifierMockGithubPullRequestGetter {
	return &VerifierMockGithubPullRequestGetter{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
		inOrderContext:         inOrderContext,
	}
}

func (mock *MockGithubPullRequestGetter) VerifyWasCalledEventually(invocationCountMatcher pegomock.InvocationCountMatcher, timeout time.Duration) *VerifierMockGithubPullRequestGetter {
	return &VerifierMockGithubPullRequestGetter{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
		timeout:                timeout,
	}
}

type VerifierMockGithubPullRequestGetter struct {
	mock                   *MockGithubPullRequestGetter
	invocationCountMatcher pegomock.InvocationCountMatcher
	inOrderContext         *pegomock.InOrderContext
	timeout                time.Duration
}

func (verifier *VerifierMockGithubPullRequestGetter) GetPullRequest(repo models.Repo, pullNum int) *MockGithubPullRequestGetter_GetPullRequest_OngoingVerification {
	params := []pegomock.Param{repo, pullNum}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "GetPullRequest", params, verifier.timeout)
	return &MockGithubPullRequestGetter_GetPullRequest_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type MockGithubPullRequestGetter_GetPullRequest_OngoingVerification struct {
	mock              *MockGithubPullRequestGetter
	methodInvocations []pegomock.MethodInvocation
}

func (c *MockGithubPullRequestGetter_GetPullRequest_OngoingVerification) GetCapturedArguments() (models.Repo, int) {
	repo, pullNum := c.GetAllCapturedArguments()
	return repo[len(repo)-1], pullNum[len(pullNum)-1]
}

func (c *MockGithubPullRequestGetter_GetPullRequest_OngoingVerification) GetAllCapturedArguments() (_param0 []models.Repo, _param1 []int) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]models.Repo, len(c.methodInvocations))
		for u, param := range params[0] {
			_param0[u] = param.(models.Repo)
		}
		_param1 = make([]int, len(c.methodInvocations))
		for u, param := range params[1] {
			_param1[u] = param.(int)
		}
	}
	return
}

func (verifier *VerifierMockGithubPullRequestGetter) GetPullRequestFromName(repoName string, repoOwner string, pullNum int) *MockGithubPullRequestGetter_GetPullRequestFromName_OngoingVerification {
	params := []pegomock.Param{repoName, repoOwner, pullNum}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "GetPullRequestFromName", params, verifier.timeout)
	return &MockGithubPullRequestGetter_GetPullRequestFromName_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type MockGithubPullRequestGetter_GetPullRequestFromName_OngoingVerification struct {
	mock              *MockGithubPullRequestGetter
	methodInvocations []pegomock.MethodInvocation
}

func (c *MockGithubPullRequestGetter_GetPullRequestFromName_OngoingVerification) GetCapturedArguments() (string, string, int) {
	repoName, repoOwner, pullNum := c.GetAllCapturedArguments()
	return repoName[len(repoName)-1], repoOwner[len(repoOwner)-1], pullNum[len(pullNum)-1]
}

func (c *MockGithubPullRequestGetter_GetPullRequestFromName_OngoingVerification) GetAllCapturedArguments() (_param0 []string, _param1 []string, _param2 []int) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(c.methodInvocations))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
		_param1 = make([]string, len(c.methodInvocations))
		for u, param := range params[1] {
			_param1[u] = param.(string)
		}
		_param2 = make([]int, len(c.methodInvocations))
		for u, param := range params[2] {
			_param2[u] = param.(int)
		}
	}
	return
}
