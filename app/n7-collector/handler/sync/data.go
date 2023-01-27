package sync

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-collector/handler/datasource"
	"github.com/eviltomorrow/project-n7/app/n7-collector/handler/db"
	"github.com/eviltomorrow/project-n7/lib/mathutil"
	"github.com/eviltomorrow/project-n7/lib/model"
	"github.com/eviltomorrow/project-n7/lib/mongodb"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	"go.uber.org/zap"
)

var (
	inFlightSem = make(chan struct{}, 1)
	CodeList    = []string{
		"sh688***",
		"sh605***",
		"sh603***",
		"sh601***",
		"sh600***",
		"sz300***",
		"sz0030**",
		"sz002***",
		"sz001***",
		"sz000***",
	}
	FetchFactories = map[string]func([]string) ([]*model.Metadata, error){
		"sina":   datasource.FetchMetadataFromSina,
		"net126": datasource.FetchMetadataFromNet126,
	}
	Size       = 30
	Limit      = 3
	Timeout    = 10 * time.Second
	RandomWait = [2]int{10, 30}
)

func DataQuick(source string) (int64, int64, error) {
	select {
	case inFlightSem <- struct{}{}:
		defer func() { <-inFlightSem }()
	default:
		return 0, 0, fmt.Errorf("sync data is busy")
	}

	var pipe = make(chan *model.Metadata, 128)
	go storeData(source, pipe)

	var options = []func(*model.Metadata) bool{
		func(m *model.Metadata) bool {
			if m != nil && m.Volume != 0 {
				return true
			}
			return false
		},
	}
	return fetchData(source, false, pipe, options)
}

func DataSlow(source string) (int64, int64, error) {
	select {
	case inFlightSem <- struct{}{}:
		defer func() { <-inFlightSem }()
	default:
		return 0, 0, fmt.Errorf("sync data is busy")
	}

	var pipe = make(chan *model.Metadata, 128)
	go storeData(source, pipe)

	var today = time.Now().Format("2006-01-02")
	var options = []func(*model.Metadata) bool{
		func(m *model.Metadata) bool {
			if m != nil && m.Date == today {
				return true
			}
			return false
		},
		func(m *model.Metadata) bool {
			if m != nil && m.Volume != 0 {
				return true
			}
			return false
		},
	}
	return fetchData(source, true, pipe, options)
}

func fetchData(source string, slow bool, pipe chan *model.Metadata, options []func(*model.Metadata) bool) (int64, int64, error) {
	defer func() {
		close(pipe)
	}()

	fetch, ok := FetchFactories[source]
	if !ok {
		return 0, 0, fmt.Errorf("not found fetchFunc, source = [%s]", source)
	}

	var (
		retrytimes                    = 0
		totalCount, ignoreCount int64 = 0, 0
		codeList                      = make([]string, 0, Size)
	)

	for code := range genCode() {
		codeList = append(codeList, code)
		if len(codeList) == Size {
		retry_1:
			data, err := fetch(codeList)
			if err != nil {
				retrytimes++
				if retrytimes == Limit {
					return totalCount, ignoreCount, fmt.Errorf("FetchMeatadata failure, nest error: %v, source: [%v], codeList: %v", err, source, codeList)
				} else {
					time.Sleep(3 * time.Minute)
					goto retry_1
				}
			}
			retrytimes = 0
			codeList = codeList[:0]

		next1:
			for _, d := range data {
				for _, f := range options {
					if ok := f(d); !ok {
						zlog.Warn("Ignore data", zap.String("data", d.String()))
						ignoreCount++
						continue next1
					}
				}

				pipe <- d
				totalCount++
			}

			if slow {
				time.Sleep(time.Duration(mathutil.GenRandInt(RandomWait[0], RandomWait[1])) * time.Second)
			} else {
				time.Sleep(300 * time.Millisecond)
			}
		}
	}

	if len(codeList) != 0 {
		retrytimes = 0
	retry_2:
		data, err := fetch(codeList)
		if err != nil {
			retrytimes++
			if retrytimes == Limit {
				return totalCount, ignoreCount, fmt.Errorf("FetchMeatadata failure, nest error: %v, source: [%v], codeList: %v", err, source, codeList)
			} else {
				time.Sleep(3 * time.Minute)
				goto retry_2
			}
		}
	next2:
		for _, d := range data {
			for _, f := range options {
				if ok := f(d); !ok {
					zlog.Warn("Ignore data", zap.String("data", d.String()))
					ignoreCount++
					continue next2
				}
			}

			pipe <- d
			totalCount++
		}
	}
	return totalCount, ignoreCount, nil
}

func storeData(source string, pipe chan *model.Metadata) {
	var dataList = make([]*model.Metadata, 0, Size)
	for data := range pipe {
		if _, err := db.DeleteMetadataByDate(mongodb.DB, source, data.Code, data.Date, Timeout); err != nil {
			zlog.Error("DeleteMetadata failure", zap.Error(err), zap.String("data", data.String()))
		} else {
			dataList = append(dataList, data)
			if len(dataList) == Size {
				if _, err := db.InsertMetadataMany(mongodb.DB, source, dataList, Timeout); err != nil {
					for _, d := range dataList {
						zlog.Error("InsertMetadata failure", zap.Error(err), zap.String("data", d.String()))
					}
				}
				dataList = dataList[:0]
			}
		}
	}

	var cache = make([]*model.Metadata, 0, len(dataList))
	if len(dataList) != 0 {
		for _, data := range dataList {
			if _, err := db.DeleteMetadataByDate(mongodb.DB, source, data.Code, data.Date, Timeout); err != nil {
				zlog.Error("DeleteMetadata failure", zap.Error(err), zap.String("data", data.String()))
			} else {
				cache = append(cache, data)
			}
		}
	}

	if len(cache) != 0 {
		if _, err := db.InsertMetadataMany(mongodb.DB, source, cache, Timeout); err != nil {
			for _, d := range cache {
				zlog.Error("InsertMetadata failure", zap.Error(err), zap.String("data", d.String()))
			}
		}
	}
}

func genCode() chan string {
	var data = make(chan string, 64)
	go func() {
		for _, code := range CodeList {
			result, err := buildCode(code)
			if err != nil {
				zlog.Error("Build range code failure", zap.Error(err))
				continue
			}
			for _, r := range result {
				data <- r
			}
		}
		close(data)
	}()
	return data
}

func buildCode(baseCode string) ([]string, error) {
	if len(baseCode) != 8 {
		return nil, fmt.Errorf("code length must be 8, code is [%s]", baseCode)
	}
	if !strings.HasPrefix(baseCode, "sh") && !strings.HasPrefix(baseCode, "sz") {
		return nil, fmt.Errorf("code must be start with [sh/sz], code is [%s]", baseCode)
	}

	if !strings.Contains(baseCode, "*") {
		return []string{baseCode}, nil
	}

	var (
		n      = strings.Index(baseCode, "*")
		prefix = baseCode[:n]
		codes  = make([]string, 0, int(math.Pow10(8-n)))
	)

	var builder strings.Builder
	builder.Grow(8)

	var next = int(math.Pow10(8-n)) - 1
	var mid = ""
	var count = -1
	var changed = false
	for i := next; i >= 0; i-- {
		if i == next && i != 0 {
			next = i / 10
			count++
			changed = true
			mid = ""
		} else {
			changed = false
		}

		if changed {
			for j := 0; j < count; j++ {
				mid += "0"
			}
		}

		builder.WriteString(prefix)
		builder.WriteString(mid)
		builder.WriteString(strconv.Itoa(i))
		codes = append(codes, builder.String())
		builder.Reset()
	}
	return codes, nil
}
