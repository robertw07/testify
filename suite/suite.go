package suite

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allTestsFilter = func(_, _ string) (bool, error) { return true, nil }
var matchMethod = flag.String("testify.m", "", "regular expression to select tests of the testify suite to run")

// Suite is a basic testing suite with methods for storing and
// retrieving the current *testing.T context.
type Suite struct {
	*assert.Assertions
	mu      sync.RWMutex
	require *require.Assertions
	t       *testing.T
}

// T retrieves the current *testing.T context.
func (suite *Suite) T() *testing.T {
	suite.mu.RLock()
	defer suite.mu.RUnlock()
	return suite.t
}

// SetT sets the current *testing.T context.
func (suite *Suite) SetT(t *testing.T) {
	suite.mu.Lock()
	defer suite.mu.Unlock()
	suite.t = t
	suite.Assertions = assert.New(t)
	suite.require = require.New(t)
}

// Require returns a require context for suite.
func (suite *Suite) Require() *require.Assertions {
	suite.mu.Lock()
	defer suite.mu.Unlock()
	if suite.require == nil {
		suite.require = require.New(suite.T())
	}
	return suite.require
}

// Assert returns an assert context for suite.  Normally, you can call
// `suite.NoError(expected, actual)`, but for situations where the embedded
// methods are overridden (for example, you might want to override
// assert.Assertions with require.Assertions), this method is provided so you
// can call `suite.Assert().NoError()`.
func (suite *Suite) Assert() *assert.Assertions {
	suite.mu.Lock()
	defer suite.mu.Unlock()
	if suite.Assertions == nil {
		suite.Assertions = assert.New(suite.T())
	}
	return suite.Assertions
}

func recoverAndFailOnPanic(t *testing.T) {
	r := recover()
	failOnPanic(t, r)
}

func failOnPanic(t *testing.T, r interface{}) {
	if r != nil {
		t.Errorf("test panicked: %v\n%s", r, debug.Stack())
		t.FailNow()
	}
}

// Run provides suite functionality around golang subtests.  It should be
// called in place of t.Run(name, func(t *testing.T)) in test suite code.
// The passed-in func will be executed as a subtes≈t with a fresh instance of t.
// Provides compatibility with go test pkg -run TestSuite/TestName/SubTestName.
func (suite *Suite) Run(name string, subtest func()) bool {
	oldT := suite.T()
	defer suite.SetT(oldT)
	return oldT.Run(name, func(t *testing.T) {
		suite.SetT(t)
		subtest()
	})
}

// Run takes a testing suite and runs all of the tests attached
// to it.
func Run(t *testing.T, suite TestingSuite, caseInfos []CaseInfo) {
	defer recoverAndFailOnPanic(t)
	suite.SetT(t)

	var suiteSetupDone bool

	var stats *SuiteInformation
	if _, ok := suite.(WithStats); ok {
		stats = newSuiteInformation()
	}

	tests := []testing.InternalTest{}
	methodFinder := reflect.TypeOf(suite)

	suiteName := methodFinder.Elem().Name()

	for i := 0; i < methodFinder.NumMethod(); i++ {
		method := methodFinder.Method(i)

		ok, err := methodFilter(method.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "testify: invalid regexp for -m: %s\n", err)
			os.Exit(1)
		}

		if !ok {
			continue
		}

		if !suiteSetupDone {
			if stats != nil {
				stats.Start = time.Now()
			}

			if setupAllSuite, ok := suite.(SetupAllSuite); ok {
				setupAllSuite.SetupSuite()
			}

			suiteSetupDone = true
		}

		filedFinder := reflect.ValueOf(suite)

		// Add by Robert for extend test features
		dataField := filedFinder.Elem().FieldByName("TestData")
		//fmt.Println(dataField)
		testData := map[string]interface{}{}
		if dataField != (reflect.Value{}) {
			mapIter := dataField.MapRange()
			for mapIter.Next() {
				if mapIter.Value().Elem().Kind() == reflect.Slice {
					testData[mapIter.Key().Interface().(string)] = mapIter.Value().Elem().Interface()
				}
			}
		}

		test := testing.InternalTest{
			method.Name,
			func(t *testing.T) {
				parentT := suite.T()
				suite.SetT(t)
				defer recoverAndFailOnPanic(t)
				defer func() {
					r := recover()

					if stats != nil {
						passed := !t.Failed() && r == nil
						stats.end(method.Name, passed)
					}

					if afterTestSuite, ok := suite.(AfterTest); ok {
						afterTestSuite.AfterTest(suiteName, method.Name)
					}

					if tearDownTestSuite, ok := suite.(TearDownTestSuite); ok {
						tearDownTestSuite.TearDownTest()
					}

					suite.SetT(parentT)
					failOnPanic(t, r)
				}()

				if setupTestSuite, ok := suite.(SetupTestSuite); ok {
					setupTestSuite.SetupTest()
				}
				if beforeTestSuite, ok := suite.(BeforeTest); ok {
					beforeTestSuite.BeforeTest(methodFinder.Elem().Name(), method.Name)
				}

				if stats != nil {
					stats.start(method.Name)
				}

				// Add by Robert for extend test features
				// test feature extension
				curCaseInfo := getCaseInfo(caseInfos, method.Name)
				if curCaseInfo != nil {
					if curCaseInfo.IsSkip {
						t.Skipf("Current test method Tags:%s", curCaseInfo.TagStr)
					}
					// Concurrent execution of data-driven testing cases
					if curCaseInfo.DataKey != "" {
						targetData := testData[curCaseInfo.DataKey]
						if reflect.ValueOf(targetData).Kind() == reflect.Slice {

							itemS := reflect.ValueOf(targetData)
							var targetDataArray []interface{}
							for i := 0; i < itemS.Len(); i++ {
								ele := itemS.Index(i)
								targetDataArray = append(targetDataArray, ele.Interface())
							}

							parallelCount := curCaseInfo.ParallelCount
							var wg sync.WaitGroup
							if parallelCount == 0 {
								parallelCount = 1
							}

							failCount := 0
							ch := make(chan struct{}, parallelCount)
							for i := 0; i < len(targetDataArray); i++ {
								curData := targetDataArray[i]
								ch <- struct{}{}
								wg.Add(1)
								caseName := fmt.Sprintf("%s_%d_%s", curCaseInfo.DataKey, i, buildSubCaseName(curData))
								go func(caseNameStr string, data interface{}) {
									t.Run(caseNameStr, func(tt *testing.T) {
										defer wg.Done()
										defer recoverAndFailOnPanic(tt)
										defer func() {
											r := recover()

											if stats != nil {
												passed := !tt.Failed() && r == nil
												stats.end(method.Name, passed)
											}

											if afterTestSuite, ok := suite.(AfterTest); ok {
												afterTestSuite.AfterTest(suiteName, method.Name)
											}

											if tearDownTestSuite, ok := suite.(TearDownTestSuite); ok {
												tearDownTestSuite.TearDownTest()
											}

											suite.SetT(parentT)
											<-ch //
											failOnPanic(tt, r)
										}()
										method.Func.Call([]reflect.Value{reflect.ValueOf(suite), reflect.ValueOf(data), reflect.ValueOf(tt)})
										if tt.Failed() {
											failCount++
										}
									})
								}(caseName, curData)
							}
							wg.Wait()

							var failureRate float64
							if len(targetDataArray) != 0 {
								failureRate = (float64(len(targetDataArray)-failCount) / float64(len(targetDataArray))) * 100
								failureRate, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", failureRate), 64)
							} else {
								failureRate = 0
							}
							t.Logf("Total data case count: %d", len(targetDataArray))
							t.Logf("Failed data case count: %d", failCount)
							t.Log("Success rate: ", failureRate, "%")
							t.Logf("Parallel count: %d", parallelCount)
						}

					} else {
						method.Func.Call([]reflect.Value{reflect.ValueOf(suite)})
					}
				} else {
					method.Func.Call([]reflect.Value{reflect.ValueOf(suite)})
				}
			},
		}
		tests = append(tests, test)
	}
	if suiteSetupDone {
		defer func() {
			if tearDownAllSuite, ok := suite.(TearDownAllSuite); ok {
				tearDownAllSuite.TearDownSuite()
			}

			if suiteWithStats, measureStats := suite.(WithStats); measureStats {
				stats.End = time.Now()
				suiteWithStats.HandleStats(suiteName, stats)
			}
		}()
	}

	runTests(t, tests)
}

func buildSubCaseName(data interface{}) string {
	caseName := ""
	dataValue := reflect.ValueOf(data)
	switch dataValue.Kind() {
	case reflect.String:
		str := data.(string)
		len0 := len(str)
		if len0 < 50 {
			caseName = str[0 : len(str)-1]
		} else {
			caseName = str[0:49] + "..."
		}
		break
	case reflect.Struct:
		fName := dataValue.FieldByName("CaseDesc")
		if fName == (reflect.Value{}) {
			fName = dataValue.FieldByName("Name")
		}
		if fName == (reflect.Value{}) {
			fName = dataValue.FieldByName("CaseName")
		}
		// for testing
		//if fName == (reflect.Value{}) {
		//	fName = dataValue.FieldByName("SuiteName")
		//}
		caseName = fName.String()
		break
	case reflect.Int, reflect.Int16, reflect.Int64:
		caseName = fmt.Sprintf("%d", dataValue.Int())
		break
	case reflect.Float32, reflect.Float64:
		caseName = fmt.Sprintf("%f", dataValue.Float())
		break
	}
	return caseName
}

func getCaseInfo(caseInfos []CaseInfo, methodName string) *CaseInfo {
	for _, caseInfo := range caseInfos {
		if caseInfo.MethodName == methodName {
			return &caseInfo
		}
	}
	return nil
}

func filterSkipCase(skipCases []string, methodName string) bool {
	for i := 0; i < len(skipCases); i++ {
		if skipCases[i] == methodName {
			return true
		}
	}
	return false
}

// Filtering method according to set regular expression
// specified command-line argument -m
func methodFilter(name string) (bool, error) {
	if ok, _ := regexp.MatchString("^Test", name); !ok {
		return false, nil
	}
	return regexp.MatchString(*matchMethod, name)
}

func runTests(t testing.TB, tests []testing.InternalTest) {
	if len(tests) == 0 {
		t.Log("warning: no tests to run")
		return
	}

	r, ok := t.(runner)
	if !ok { // backwards compatibility with Go 1.6 and below
		if !testing.RunTests(allTestsFilter, tests) {
			t.Fail()
		}
		return
	}

	for _, test := range tests {
		r.Run(test.Name, test.F)
	}
}

type runner interface {
	Run(name string, f func(t *testing.T)) bool
}
