package planner

import (
	"math"
)

// ForecastDay represents a single day's forecast data
type ForecastDay struct {
	Date     string
	TempMax  float64
	TempMin  float64
	PrecipMm float64
}

// PlannerConfig contains algorithm configuration
type PlannerConfig struct {
	SoilOptimalMin      int     // 土壤最优湿度下限 (ADC值)
	SoilOptimalMax      int     // 土壤最优湿度上限 (ADC值)
	MaxIrrigationPerDay float64 // 每日最大灌溉量 (升)
	BaseET              float64 // 基础蒸散量 (升/天)
	TempFactor          float64 // 温度系数
	RainConversion      float64 // 降雨转换系数
	ADCToMoisture       float64 // ADC到湿度的转换系数
	CostW1              float64 // 代价权重1 (湿度偏差)
	CostW2              float64 // 代价权重2 (用水量)
	CostW3              float64 // 代价权重3 (变化平滑)
}

// DailyPlan represents the irrigation plan for one day
type DailyPlan struct {
	Date           string
	PlannedVolumeL float64
}

// IrrigationPlanner implements the DP-based irrigation planning algorithm
type IrrigationPlanner struct {
	config PlannerConfig
}

// NewIrrigationPlanner creates a new planner instance
func NewIrrigationPlanner(config PlannerConfig) *IrrigationPlanner {
	return &IrrigationPlanner{config: config}
}

// ComputePlan computes the optimal irrigation plan using dynamic programming
func (p *IrrigationPlanner) ComputePlan(initialSoilMoisture int, forecasts []ForecastDay) []DailyPlan {
	if len(forecasts) == 0 {
		return []DailyPlan{}
	}

	days := len(forecasts)

	// 离散化灌溉量选项：0, 0.5, 1.0, 1.5, ..., MaxIrrigationPerDay
	maxSteps := int(p.config.MaxIrrigationPerDay / 0.5)
	irrigationOptions := make([]float64, maxSteps+1)
	for i := range irrigationOptions {
		irrigationOptions[i] = float64(i) * 0.5
	}

	// DP状态: dp[day][moisture] = {cost, prevMoisture, irrigationVolume}
	// moisture范围: 0-4095 (ADC范围)
	// 为了减少状态空间，按步长10离散化
	moistureStep := 10
	maxMoisture := 4095
	moistureStates := maxMoisture/moistureStep + 1

	type State struct {
		cost             float64
		prevMoisture     int
		irrigationVolume float64
	}

	// 初始化DP表
	dp := make([][]State, days+1)
	for i := range dp {
		dp[i] = make([]State, moistureStates)
		for j := range dp[i] {
			dp[i][j].cost = math.Inf(1) // 无穷大
		}
	}

	// 初始状态
	initMoistureIdx := initialSoilMoisture / moistureStep
	if initMoistureIdx >= moistureStates {
		initMoistureIdx = moistureStates - 1
	}
	dp[0][initMoistureIdx].cost = 0
	dp[0][initMoistureIdx].prevMoisture = initialSoilMoisture
	dp[0][initMoistureIdx].irrigationVolume = 0

	optimalCenter := float64(p.config.SoilOptimalMin+p.config.SoilOptimalMax) / 2.0

	// DP转移
	for day := 0; day < days; day++ {
		forecast := forecasts[day]

		// 计算蒸散量
		tAvg := (forecast.TempMax + forecast.TempMin) / 2.0
		et := p.config.BaseET + p.config.TempFactor*(tAvg-20.0)

		// 降雨补充 (毫米转换为湿度增量)
		rainMoisture := forecast.PrecipMm * p.config.RainConversion * p.config.ADCToMoisture

		for currIdx := 0; currIdx < moistureStates; currIdx++ {
			if math.IsInf(dp[day][currIdx].cost, 1) {
				continue
			}

			currMoisture := currIdx * moistureStep

			// 尝试所有灌溉选项
			for _, irrigation := range irrigationOptions {
				// 湿度转移方程
				newMoisture := float64(currMoisture) - et*p.config.ADCToMoisture + irrigation*p.config.ADCToMoisture + rainMoisture

				// 边界约束
				if newMoisture < 0 {
					newMoisture = 0
				}
				if newMoisture > float64(maxMoisture) {
					newMoisture = float64(maxMoisture)
				}

				newIdx := int(newMoisture) / moistureStep
				if newIdx >= moistureStates {
					newIdx = moistureStates - 1
				}

				// 计算代价
				moistureDeviation := math.Abs(newMoisture - optimalCenter)
				prevIrrigation := dp[day][currIdx].irrigationVolume
				irrigationChange := math.Abs(irrigation - prevIrrigation)

				cost := p.config.CostW1*moistureDeviation*moistureDeviation +
					p.config.CostW2*irrigation +
					p.config.CostW3*irrigationChange

				totalCost := dp[day][currIdx].cost + cost

				// 更新状态
				if totalCost < dp[day+1][newIdx].cost {
					dp[day+1][newIdx].cost = totalCost
					dp[day+1][newIdx].prevMoisture = currMoisture
					dp[day+1][newIdx].irrigationVolume = irrigation
				}
			}
		}
	}

	// 回溯最优路径
	// 找到最后一天代价最小的状态
	minCost := math.Inf(1)
	bestIdx := 0
	for idx := 0; idx < moistureStates; idx++ {
		if dp[days][idx].cost < minCost {
			minCost = dp[days][idx].cost
			bestIdx = idx
		}
	}

	// 构建计划
	plan := make([]DailyPlan, days)
	currentIdx := bestIdx

	// 回溯路径
	for day := days - 1; day >= 0; day-- {
		plan[day] = DailyPlan{
			Date:           forecasts[day].Date,
			PlannedVolumeL: dp[day+1][currentIdx].irrigationVolume,
		}

		// 回到前一天的状态
		if day > 0 {
			prevMoisture := dp[day+1][currentIdx].prevMoisture
			currentIdx = prevMoisture / moistureStep
		}
	}

	return plan
}
