// Assert response code is correct:
//    Then the response code should be 202
// Assert that a json field of the response body is correct.  This uses a http://github.com/itchyny/gojq expression to select the json field of the
// response:
//    Then the ".status" selection from the response should match "assigning"
// Assert that the response body matches the provided text:
//    Then the response should match "Hello"
// Assert that response json matches the provided json.  Differences in json formatting and field order are ignored.:
//    Then the response should match json:
//      """
//      {
//          "id": "${cid}",
//      }
//      """
// Stores a json field of the response body in a scenario variable:
//    Given I store the ".id" selection from the response as ${cid}
//
package cucumber

import (
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/itchyny/gojq"
	"reflect"
)

func init() {
	StepModules = append(StepModules, func(ctx *godog.ScenarioContext, s *TestScenario) {
		ctx.Step(`^the response code should be (\d+)$`, s.theResponseCodeShouldBe)
		ctx.Step(`^the response should match json:$`, s.theResponseShouldMatchJsonDoc)
		ctx.Step(`^the response should match "([^"]*)"$`, s.theResponseShouldMatchText)
		ctx.Step(`^I store the "([^"]*)" selection from the response as \${([^"]*)}$`, s.iStoreTheSelectionFromTheResponseAs)
		ctx.Step(`^the "([^"]*)" selection from the response should match "([^"]*)"$`, s.theSelectionFromTheResponseShouldMatch)
	})
}

func (s *TestScenario) theResponseCodeShouldBe(expected int) error {
	session := s.Session()
	actual := session.Resp.StatusCode
	if expected != actual {
		return fmt.Errorf("expected response code to be: %d, but actual is: %d, body: %s", expected, actual, string(session.RespBytes))
	}
	return nil
}
func (s *TestScenario) theResponseShouldMatchJsonDoc(expected *godog.DocString) error {
	return s.theResponseShouldMatchJson(expected.Content)
}

func (s *TestScenario) theResponseShouldMatchJson(expected string) error {
	session := s.Session()
	var expectedParsed interface{}

	expanded := s.Expand(expected)
	if err := json.Unmarshal([]byte(expanded), &expectedParsed); err != nil {
		return fmt.Errorf("expected JSON is not valid json: %v, expected: %s", err, expanded)
	}

	// parse both so we can deep compare...
	actualParsed, err := session.RespJson()
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(expectedParsed, actualParsed) {
		expected, _ := json.MarshalIndent(expectedParsed, "", "  ")
		actual, _ := json.MarshalIndent(actualParsed, "", "  ")
		return fmt.Errorf("reponse does not match\nexpected:\n%v\nactual:\n%v", string(expected), string(actual))
	}
	return nil
}

func (s *TestScenario) theResponseShouldMatchText(expected string) error {
	session := s.Session()

	expanded := s.Expand(expected)

	if expanded != string(session.RespBytes) {
		return fmt.Errorf("reponse does not match expected: %v, actual: %v", expanded, string(session.RespBytes))
	}
	return nil
}

func (s *TestScenario) iStoreTheSelectionFromTheResponseAs(selector string, as string) error {

	session := s.Session()
	doc, err := session.RespJson()
	if err != nil {
		return err
	}

	query, err := gojq.Parse(selector)
	if err != nil {
		return err
	}

	iter := query.Run(doc)
	if next, found := iter.Next(); found {
		s.Variables[as] = fmt.Sprintf("%s", next)
		return nil
	}
	return fmt.Errorf("expected JSON does not have node that matches selector: %s", selector)
}

func (s *TestScenario) theSelectionFromTheResponseShouldMatch(selector string, expected string) error {
	session := s.Session()
	doc, err := session.RespJson()
	if err != nil {
		return err
	}

	query, err := gojq.Parse(selector)
	if err != nil {
		return err
	}

	expected = s.Expand(expected)
	iter := query.Run(doc)
	if actual, found := iter.Next(); found {
		actual := fmt.Sprintf("%v", actual)
		if actual != expected {
			return fmt.Errorf("selected JSON does not match. expected: %v, actual: %v", expected, actual)
		}
		return nil
	}
	return fmt.Errorf("expected JSON does not have node that matches selector: %s", selector)
}