package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DemoSuiteTest struct {
	suite.Suite
	TestData map[string][]string
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
	}
	suite.Run(t, new(DemoSuiteTest), caseInfos)
}

func (t *DemoSuiteTest) SetupSuite() {
	t.TestData = map[string][]string{}
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
