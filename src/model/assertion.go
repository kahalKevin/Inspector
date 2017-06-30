package model

// Define our message object
type AssertionResult struct {
	PyScript      string                 `json:"pyscript"`
	Title         string                 `json:"title"`
	Times   	  int   		         `json:"times"`
	Cleared       bool                   `json:"cleared"`
    Asserted      AssertionEntities      `json:"entity"`
    EnvEnts  	  EnvMonitoringEntities  `json:"envent"`
}

type AssertionResults []AssertionResult

type AssertionEntity struct {
	Key      string `json:"key"`
	Success  bool   `json:"value"`
	Info     string `json:"info"`
}

type AssertionEntities []AssertionEntity

type EnvMonitoringEntity struct{
	Env   	 string   `json:"env"`
	Free   	 float64  `json:"free"`
	Info     string   `json:"info"`
}

type EnvMonitoringEntities []EnvMonitoringEntity