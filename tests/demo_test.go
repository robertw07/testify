package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DemoSuiteTest struct {
	suite.Suite
	TestData map[string]interface{}
}

func TestDemo(t *testing.T) {
	caseInfos := []suite.CaseInfo{
		{
			MethodName:    "Test_1",
			DataKey:       "EthTop200",
			ParallelCount: 1,
		},
		{
			MethodName: "Test_0",
			IsSkip:     true,
		},
		{
			MethodName:    "Test_2",
			DataKey:       "Test",
			ParallelCount: 1,
		},
		{
			MethodName:    "Test_3",
			DataKey:       "Test",
			ParallelCount: 1,
		},
		{
			MethodName:    "Test_4",
			DataKey:       "Test1",
			ParallelCount: 1,
		},
		{
			MethodName:    "Test_5",
			DataKey:       "Test2",
			ParallelCount: 1,
		},
	}
	suite.Run(t, new(DemoSuiteTest), caseInfos)
}

func (t *DemoSuiteTest) SetupSuite() {
	t.TestData = map[string]interface{}{}
	//t.DataOperator.GetOnlineData("....")
	t.TestData["EthTop200"] = []string{
		"12113241234123",
		"abcasdfasdfasd",
		"13241234123412",
		"************",
		"777777777777",
		"122333333333",
		"************",
		"777777777777",
	}
	t.TestData["EthTop500"] = []string{
		"************",
		"777777777777",
		"122333333333",
	}

	t.TestData["Test"] = []suite.CaseInfo{{SuiteName: "abc"}, {SuiteName: "bcd"}}
	t.TestData["Test1"] = []float64{1.001, 2.001, 3.001}
	t.TestData["Test2"] = []int{1, 2, 3}
	//t.TestData["EthTop200"] = t.DataOperator.ReadFileLines("/***.txt")
	//t.TestData["EthTop50"] = t.DataOperator.ReadFileLines("")
}

func (t *DemoSuiteTest) Test_0() {
	fmt.Println("*****")
}

func (t *DemoSuiteTest) Test_1(data string, tt *testing.T) {
	tt.Log("~~~~:", data)

	if data == "777777777777" {
		a := []string{}
		fmt.Sprintf(a[3])
		assert.True(tt, false, "")
	}
	//assert.True(tt, false, data)
	//time.Sleep(time.Second * 3)
}

func (t *DemoSuiteTest) Test_2(data int, tt *testing.T) {
	tt.Log("~~~~:", data)

	if data == 1 {
		a := []string{}
		fmt.Sprintf(a[3])
		assert.True(tt, false, "")
	}
}

func (t *DemoSuiteTest) Test_3(data suite.CaseInfo, tt *testing.T) {
	tt.Log("~~~~:", data)

	if data.SuiteName == "abc" {
		a := []string{}
		fmt.Sprintf(a[3])
		assert.True(tt, false, "")
	}
}

func (t *DemoSuiteTest) Test_4(data float64, tt *testing.T) {
	tt.Log("~~~~:", data)

	if data == 1 {
		a := []string{}
		fmt.Sprintf(a[3])
		assert.True(tt, false, "")
	}
}

func (t *DemoSuiteTest) Test_5(data int, tt *testing.T) {
	tt.Log("~~~~:", data)

	if data == 1 {
		a := []string{}
		fmt.Sprintf(a[3])
		assert.True(tt, false, "")
	}
}

func (t *DemoSuiteTest) Test_6(data int, tt *testing.T) {
	arry := []string{"1"}
	fmt.Println(arry[2])
}
