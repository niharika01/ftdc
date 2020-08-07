package ftdc

import (
	//"bytes"
	//"compress/zlib"
	"context"
	"encoding/csv"
	"fmt"
	//"strings"

	"github.com/evergreen-ci/birch/bsontype"
	"github.com/pkg/errors"
	"io"
	//"os"
	"strconv"
	"time"
)

func (c *Chunk) getFTDCFieldNames(operationName string) []string {
	//Field names are [ts id counters.n counters.ops counters.size counters.errors timers.dur timers.total gauges.state gauges.workers gauges.failed]
	fieldNames := make([]string, len(c.Metrics))
	for idx, m := range c.Metrics {
		fieldNames[idx] = m.Key()
	}
	//return only the field names required by t2- timestamp,counters.ops,counters.size,timers.dur,timers.total
	prefix:="cedar\000"+operationName+"\000"
	requiredFields:=[]string{prefix+fieldNames[0],prefix+fieldNames[3],prefix+fieldNames[4],prefix+fieldNames[6],prefix+fieldNames[7]}

	return requiredFields
}





func (c *Chunk) getGennyRecord(i int) []string {
	fields := make([]string, len(c.Metrics))

	for idx, m := range c.Metrics {
		switch m.originalType {
		case bsontype.Double, bsontype.Int32, bsontype.Int64, bsontype.Boolean, bsontype.Timestamp:

			fields[idx] = strconv.FormatInt(m.Values[i], 10)

		case bsontype.DateTime:

			//fields[idx]= time.Unix(m.Values[i]/1000000000, 0).Format(time.RFC3339)
			fields[idx]=strconv.FormatInt(m.Values[i],10)

		}
	}

	return fields
}


//partition into windows
func partitionWindow(startValue *int64,windows*[][]string,record[] string) []string{
	//function to partition 'ts' fields in genny records to time frames of 1s

	//using timers.total as timestamp
	//offset_ns is unix time 2020-07-01T00:00:00+00:00 in RFC 3339
	var offset_ns int64=1593561600000000000
	var timestamp,err=strconv.ParseInt(record[7],10,64)
	timestamp+=offset_ns
	var strArr[]string

	if err != nil {
		fmt.Printf("Error in converting timestamp to int64")
	}
	if *startValue == -1 {
			//on the first iteration, startValue is set to the first record's timestamp
			*startValue = timestamp
	}

		if timestamp <= *startValue+1000000000 {
			//if timestamp value is leq to startValue plus 1000000000ns ~ 1s
			//add to current window
			*windows = append(*windows, record)

		} else {
			//print out the window
			//dereference the current window
			dref := *windows

			// retrieve cumulative metrics from the last record in the window
			dur := dref[len(dref)-1][6]
			total := dref[len(dref)-1][7]
			ops := dref[len(dref)-1][3]
			size := dref[len(dref)-1][4]
			lastTimestamp, _ := strconv.ParseInt(dref[len(dref)-1][7], 10, 64)
			//add the offset_ns to bring timestamp to JULY 1 2020
			offsetTimestamp:=lastTimestamp+offset_ns
			unixTimestamp := time.Unix(offsetTimestamp/1000000000,offsetTimestamp%1000000000).Format(time.RFC3339)

			strArr = []string{unixTimestamp, ops, size, dur, total}


			//change the start value, set the window to nil and start a new window
			*startValue, _ = strconv.ParseInt(record[7], 10, 64)
			*startValue+=offset_ns
			*windows = nil
			*windows = append(*windows, record)

		}



	return strArr


}
// WriteCSV exports the contents of a stream of chunks as CSV. Returns
// an error if the number of metrics changes between points, or if
// there are any errors writing data.
func WriteGennyCSV(ctx context.Context, iter *ChunkIterator, writer io.Writer,operationName string) error {
	var numFields int
	var windows [][]string

	var startValue int64=-1


	csvw := csv.NewWriter(writer)
	for iter.Next() {
		if ctx.Err() != nil {
			return errors.New("operation aborted")
		}
		chunk := iter.Chunk()
		if numFields == 0 {
			//append genny operation name to fields
			fieldNames := chunk.getFTDCFieldNames(operationName)
			if err := csvw.Write(fieldNames); err != nil {
				return errors.Wrap(err, "problem writing field names")
			}
			numFields = len(fieldNames)
		}

		for i := 0; i < chunk.nPoints; i++ {

			record := chunk.getGennyRecord(i)
			arr := partitionWindow(&startValue, &windows, record)

			if (len(arr) != 0) {
				if err := csvw.Write(arr); err != nil {
					return errors.Wrapf(err, "problem writing csv record %d of %d", i, chunk.nPoints)
				}
			}
		}
		csvw.Flush()
		if err := csvw.Error(); err != nil {
			return errors.Wrapf(err, "problem flushing csv data")
		}
	}
	if err := iter.Err(); err != nil {
		return errors.Wrap(err, "problem reading chunks")
	}

	return nil
}





