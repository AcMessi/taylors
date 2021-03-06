package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sort"
	"sync"
	"taylors/dao"
	"taylors/model"
	"taylors/model/param"
)

type stockAnalysisService struct {
}

func (analy *stockAnalysisService) AnalysisList(filter *param.AnalysisListParam) (stockList []*model.Stock, total int, err error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 1
	}
	filter.Page = 0
	filter.PageSize = 0
	filterBys, err := json.Marshal(filter)
	if err != nil {
		return
	}
	hash := md5.New()
	hash.Write(filterBys)
	md5 := hex.EncodeToString(hash.Sum(nil))

	exitCache := false
	analysisModelCacheList := NewAnalysisCache().Get(md5)

	if len(analysisModelCacheList) > 0 {
		exitCache = true
	}

	if !exitCache {
		analysisModelCacheList = analy.Calculate(filter, md5)
	}

	endIndex := page * pageSize
	startIndex := (page - 1) * pageSize
	if endIndex > len(analysisModelCacheList)-1 {
		endIndex = len(analysisModelCacheList) - 1
	}

	codes := []string{}
	for i := startIndex; i <= endIndex; i++ {
		codes = append(codes, analysisModelCacheList[i].code)
	}

	total = len(analysisModelCacheList)

	for _, code := range codes {
		stockPO, err := dao.StockDao.FindLastByCode(code)
		if err != nil {
			stockList = nil
			break
		}

		stockPO.MarketCapital /= 100000000
		stockList = append(stockList, stockPO)
	}
	return
}

func (analy *stockAnalysisService) Calculate(filter *param.AnalysisListParam, md5 string) (scoreList []*analysisModel) {
	stockCodeList, err := dao.StockDao.CodeList()
	if err != nil {
		return
	}
	scoreList = make([]*analysisModel, 0, 5000)
	analysisCh := make(chan *analysisModel, 100)

	go func() {
		for analysis := range analysisCh {
			if analysis != nil && analysis.code != "" {
				scoreList = append(scoreList, analysis)
			}
		}
	}()

	wg := sync.WaitGroup{}

	for _, stock := range stockCodeList {
		wg.Add(1)
		go func(stockObj *model.Stock) {
			defer wg.Done()
			stockPOList, err := dao.StockDao.FindByAnalysisFilter(stockObj.Code, filter.StartTime, filter.EndTime)
			if err != nil {
				return
			}

			if len(stockPOList) == 0 {
				return
			}
			if !analy.SearchFilter(stockPOList, filter) {
				return
			}

			calculateScore := analy.CalculateScore(stockPOList)
			analysisCh <- &analysisModel{
				score: calculateScore,
				code:  stockObj.Code,
			}
		}(stock)
	}
	wg.Wait()

	sort.Sort(analysisModelList(scoreList))

	NewAnalysisCache().Cache(md5, scoreList)
	return
}

func (*stockAnalysisService) CalculateScore(stockList []*model.Stock) (score float32) {
	if stockList == nil || len(stockList) == 0 {
		return
	}

	for _, stock := range stockList {
		score += float32(stock.Percent)
	}

	return
}

func (*stockAnalysisService) SearchFilter(stockList []*model.Stock, filter *param.AnalysisListParam) (isAdd bool) {
	//条件满足最小天数
	day := filter.DayMin

	currentMaxCount := 0
	currentMinCount := 0
	volumeRatioMaxCount := 0
	volumeRatioMinCount := 0
	percentMaxCount := 0
	percentMinCount := 0
	marketCapitalMaxCount := 0
	marketCapitalMinCount := 0

	for _, stockPO := range stockList {
		if filter.CurrentMax != 0 {
			if stockPO.Current < filter.CurrentMax {
				currentMaxCount++
			}
		}
		if filter.CurrentMin != 0 {
			if stockPO.Current > filter.CurrentMin {
				currentMinCount++
			}
		}

		if filter.VolumeRatioMax != 0 {
			if stockPO.VolumeRatio < filter.VolumeRatioMax {
				volumeRatioMaxCount++
			}
		}
		if filter.VolumeRatioMin != 0 {
			if stockPO.VolumeRatio > filter.VolumeRatioMin {
				volumeRatioMinCount++
			}
		}

		if filter.PercentMax != 0 {
			if stockPO.Percent < filter.PercentMax {
				percentMaxCount++
			}
		}
		if filter.PercentMin != 0 {
			if stockPO.Percent > filter.PercentMin {
				percentMinCount++
			}
		}

		if filter.MarketCapitalMax != 0 {
			if stockPO.MarketCapital < filter.MarketCapitalMax {
				marketCapitalMaxCount++
			}
		}
		if filter.MarketCapitalMin != 0 {
			if stockPO.MarketCapital > filter.MarketCapitalMin {
				marketCapitalMinCount++
			}
		}
	}

	if filter.CurrentMax != 0 && currentMaxCount < day {
		return
	}
	if filter.CurrentMin != 0 && currentMinCount < day {
		return
	}
	if filter.VolumeRatioMax != 0 && volumeRatioMaxCount < day {
		return
	}
	if filter.VolumeRatioMin != 0 && volumeRatioMinCount < day {
		return
	}
	if filter.PercentMax != 0 && percentMaxCount < day {
		return
	}
	if filter.PercentMin != 0 && percentMinCount < day {
		return
	}
	if filter.MarketCapitalMax != 0 && marketCapitalMaxCount < day {
		return
	}
	if filter.MarketCapitalMin != 0 && marketCapitalMinCount < day {
		return
	}

	isAdd = true

	return
}
