package ftdc

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/evergreen-ci/birch"
	"github.com/evergreen-ci/birch/bsontype"
	"github.com/pkg/errors"
	"io"
	"os"
	"strconv"
	"time"
)

func (c *Chunk) getFieldNames() []string {
	fieldNames := make([]string, len(c.Metrics))
	for idx, m := range c.Metrics {
		fieldNames[idx] = m.Key()
	}
	fmt.Println("Field names are", fieldNames)
	return fieldNames
}

func (c *Chunk) getRecord(i int) []string {
	fields := make([]string, len(c.Metrics))
	for idx, m := range c.Metrics {
		switch m.originalType {
		case bsontype.Double, bsontype.Int32, bsontype.Int64, bsontype.Boolean, bsontype.Timestamp:
			fields[idx] = strconv.FormatInt(m.Values[i], 10)
		case bsontype.DateTime:

			fields[idx] = time.Unix(m.Values[i]/1000, 0).Format(time.RFC3339)
		}
	}
	return fields
}

func (c *Chunk) getGennyRecord(i int) []string {
	fields := make([]string, len(c.Metrics))
	for idx, m := range c.Metrics {
		switch m.originalType {
		case bsontype.Double, bsontype.Int32, bsontype.Int64, bsontype.Boolean, bsontype.Timestamp:
			fields[idx] = strconv.FormatInt(m.Values[i], 10)
		case bsontype.DateTime:
			var offset_ns int64=1593561600000000000
			//fmt.Println("value is",m.Values[i])
			m.Values[i]=m.Values[i]+offset_ns
			//fmt.Println("value is--after",m.Values[i])
			//fields[idx]= time.Unix(m.Values[i]/1000000000, 0).Format(time.RFC3339)
			fields[idx]=strconv.FormatInt(m.Values[i],10)

		}
	}
	return fields
}

//FindAverage
func findAverage(arr[][]string){
	for _, value := range arr {
		fmt.Println("value is",value[0])


	}
}
func getColumnAvg(columnID int,multisplice [][]string) int64 {
	var sum int64 =0
	var count int64=0
	for i := 0; i < len(multisplice); i++ {
		var value,err=strconv.ParseInt(multisplice[i][columnID],10,64)
		count++
		if err==nil {
			sum += value
		}
		// do something
	}//for loop
	
	return sum/count

}

//partition into windows
func partitionWindow(startValue *int64,windows*[][]string,record[] string) []string{
	//function to partition 'ts' fields in genny records to time frames of 100ns
	//Field names are [ts id counters.n counters.ops counters.size counters.errors timers.dur timers.total gauges.state gauges.workers gauges.failed] //f

	var timestamp,err=strconv.ParseInt(record[0],10,64)
	var strArr[]string
	//strArr := make([]string,len(*windows))
	if err != nil {
		fmt.Printf("Error in converting timestamp to int64")
	}

	//1593561600000999601
	//1593561600001001374
	if *startValue==-1{
		//on the first iteration, startValue is set to the first record's timestamp
		*startValue=timestamp
	}

	if timestamp<=*startValue+100{
		//add to current window
		*windows=append(*windows,record)


	}else {
		//print out the window
		//find avg. dur
		//find avg latency -> avg dur of a window
		//find avg total for a window
		//find throughput ops/s for a window -> mention the name of the operation ex- insert throughput
		//fmt.Println("ELSE BLOCK")
		dref:=*windows
		avgDur:=getColumnAvg(6,dref)
		avgTotal:=getColumnAvg(7,dref)

		var startOps,endOps,startBytes,endBytes int64
		startOps,err=strconv.ParseInt(dref[0][3],10,64)
		endOps,err=strconv.ParseInt(dref[len(dref)-1][3],10,64)
		startBytes,err=strconv.ParseInt(dref[0][4],10,64)

		endBytes,err=strconv.ParseInt(dref[len(dref)-1][4],10,64)

		throughputOps :=endOps-startOps
		throughputBytes:=endBytes-startBytes

		*startValue,err=strconv.ParseInt(record[0],10,64)

		//write out the completed window
		/*strArr=append(strArr,strconv.FormatInt(throughputOps,10) )
		strArr=append(strArr,strconv.FormatInt(throughputBytes,10))
		strArr=append(strArr,strconv.FormatInt(avgDur,10))
		strArr=append(strArr,strconv.FormatInt(avgTotal,10))*/
		strArr=[]string{strconv.FormatInt(throughputOps,10),strconv.FormatInt(throughputBytes,10),strconv.FormatInt(avgDur,10),strconv.FormatInt(avgTotal,10)}
		//fmt.Println("strArr is",strArr)
		//encodeData:=""+strconv.FormatInt(throughputOps,10)+","+strconv.FormatInt(throughputBytes,10)+","+strconv.FormatInt(avgDur,10)+strconv.FormatInt(avgTotal,10)+""
		//compress and encode
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		//w.Write([]byte(encodeData))
		w.Close()
		//fmt.Println(b.Bytes())

		//to read that data
		r, err := zlib.NewReader(&b)

		io.Copy(os.Stdout, r)
		r.Close()
		if err!=nil{
			fmt.Println("error")
		}
		//set window to nil
		*windows=nil

		//new window starts here
		*windows=append(*windows,record)






		//var col1Slice []int = dref[:][0]
		//fmt.Println("col1",col1)
		//fmt.Println("HWLLO",dref[0][1])
		//avgDur:=findAverage(




		/*for _, value := range *windows {
			fmt.Println("value is",value[0])


		} */


	}
	//fmt.Println("before returning strArr is",strArr)

	return strArr


}
// WriteCSV exports the contents of a stream of chunks as CSV. Returns
// an error if the number of metrics changes between points, or if
// there are any errors writing data.
func WriteCSV(ctx context.Context, iter *ChunkIterator, writer io.Writer) error {
	var numFields int
	var windows [][]string
	//var arr[]string
	//var windows [][]string
	var startValue int64=-1
	//var startValue int64:=0
	fmt.Println("called")
	csvw := csv.NewWriter(writer)
	for iter.Next() {
		if ctx.Err() != nil {
			return errors.New("operation aborted")
		}
		chunk := iter.Chunk()
		if numFields == 0 {
			fieldNames := chunk.getFieldNames()
			if err := csvw.Write(fieldNames); err != nil {
				return errors.Wrap(err, "problem writing field names")
			}
			numFields = len(fieldNames)
		} else if numFields != len(chunk.Metrics) {
			return errors.New("unexpected schema change detected")
		}

		for i := 0; i < chunk.nPoints; i++ {
			//var windows [][]string
			//record := chunk.getRecord(i)
			record := chunk.getGennyRecord(i)
			arr := partitionWindow(&startValue, &windows, record)
			//fmt.Println("here the arr is",arr)
			//must return a string array
			//if err := csvw.Write(record); err != nil {
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

func getCSVFile(prefix string, count int) (io.WriteCloser, error) {
	fn := fmt.Sprintf("%s.%d.csv", prefix, count)
	writer, err := os.Create(fn)
	if err != nil {
		return nil, errors.Wrapf(err, "provlem opening file %s", fn)
	}
	return writer, nil
}

// DumpCSV writes a sequence of chunks to CSV files, creating new
// files if the iterator detects a schema change, using only the
// number of fields in the chunk to detect schema changes. DumpCSV
// writes a header row to each file.
//
// The file names are constructed as "prefix.<count>.csv".
func DumpCSV(ctx context.Context, iter *ChunkIterator, prefix string) error {
	var (
		err       error
		writer    io.WriteCloser
		numFields int
		fileCount int
		csvw      *csv.Writer
	)
	for iter.Next() {
		if ctx.Err() != nil {
			return errors.New("operation aborted")
		}

		if writer == nil {
			writer, err = getCSVFile(prefix, fileCount)
			if err != nil {
				return errors.WithStack(err)
			}
			csvw = csv.NewWriter(writer)
			fileCount++
		}

		chunk := iter.Chunk()
		if numFields == 0 {
			fieldNames := chunk.getFieldNames()
			if err = csvw.Write(fieldNames); err != nil {
				return errors.Wrap(err, "problem writing field names")
			}
			numFields = len(fieldNames)
		} else if numFields != len(chunk.Metrics) {
			if err = writer.Close(); err != nil {
				return errors.Wrap(err, "problem flushing and closing file")
			}

			writer, err = getCSVFile(prefix, fileCount)
			if err != nil {
				return errors.WithStack(err)
			}

			csvw = csv.NewWriter(writer)
			fileCount++

			// now dump header
			fieldNames := chunk.getFieldNames()
			if err := csvw.Write(fieldNames); err != nil {
				return errors.Wrap(err, "problem writing field names")
			}
			numFields = len(fieldNames)
		}

		for i := 0; i < chunk.nPoints; i++ {
			record := chunk.getRecord(i)
			if err := csvw.Write(record); err != nil {
				return errors.Wrapf(err, "problem writing csv record %d of %d", i, chunk.nPoints)
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

	if writer == nil {
		return nil
	}
	if err := writer.Close(); err != nil {
		return errors.Wrap(err, "problem writing files to disk")

	}
	return nil
}

// ConvertFromCSV takes an input stream and writes ftdc compressed
// data to the provided output writer.
//
// If the number of fields changes in the CSV fields, the first field
// with the changed number of fields becomes the header for the
// subsequent documents in the stream.
func ConvertFromCSV(ctx context.Context, bucketSize int, input io.Reader, output io.Writer) error {
	csvr := csv.NewReader(input)

	header, err := csvr.Read()
	if err != nil {
		return errors.Wrap(err, "problem reading error")
	}

	collector := NewStreamingDynamicCollector(bucketSize, output)

	defer func() {
		if err != nil && (errors.Cause(err) != context.Canceled || errors.Cause(err) != context.DeadlineExceeded) {
			err = errors.Wrap(err, "omitting final flush, because of prior error")
		}
		err = FlushCollector(collector, output)
	}()

	var record []string
	for {
		if ctx.Err() != nil {
			// this is weird so that the defer can work
			err = errors.Wrap(err, "operation aborted")
			return err
		}

		record, err = csvr.Read()
		if err == io.EOF {
			// this is weird so that the defer can work
			err = nil
			return err
		}

		if err != nil {
			if pr, ok := err.(*csv.ParseError); ok && pr.Err == csv.ErrFieldCount {
				header = record
				continue
			}
			err = errors.Wrap(err, "problem parsing csv")
			return err
		}
		if len(record) != len(header) {
			return errors.New("unexpected field count change")
		}

		elems := make([]*birch.Element, 0, len(header))
		for idx := range record {
			var val int
			val, err = strconv.Atoi(record[idx])
			if err != nil {
				continue
			}
			elems = append(elems, birch.EC.Int64(header[idx], int64(val)))
		}

		if err = collector.Add(birch.NewDocument(elems...)); err != nil {
			return errors.WithStack(err)
		}
	}
}
