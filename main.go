// Author J. Luis Cardenas.
// Use of this source code is governed by a BSD-style

package main

import (
    "fmt"
    "os"
    "io"
    "encoding/csv"
    "encoding/hex"
    "crypto/md5"
    "os/exec"
    "strings"

)

func ParseCsv(filename string, headers bool) ([][]string, error) {
  file, err := os.Open(filename)
  if err != nil {
    fmt.Println("Error:", err)
    return nil,err
  }
  defer file.Close()
  
  reader := csv.NewReader(file)
  
  reader.Comma = ','
  reader.LazyQuotes = true
  reader.TrailingComma = true
  
  var output [][]string
  for {
    record, err := reader.Read()
    if err == io.EOF {
      break
    } else if err != nil {
      return output,err;
    }
	output = append(output, record);
	
    if headers {
      break;
    }
  }

  return output,nil
}

func WriteCsv(filename string, lines [][]string) (error) {
  file, err := os.Create(filename)
  if err != nil {
    fmt.Println("Error:", err)
    return err
  }
  defer file.Close()

  writer := csv.NewWriter(file)

  writer.WriteAll(lines)

  if err := writer.Error(); err != nil {
    fmt.Println("Error writing csv", err)
    return err
  }

  return nil
}

func Exists(name string) bool {
    if _, err := os.Stat(name); err != nil {
    	if os.IsNotExist(err) {
        	return false
        }
    }
    return true
}
func IndexOf (values []string, value string) int {
    for i:=0;i<len(values);i++ {
        if (values[i] == value) {
            return i
        }
    }
    return -1
}

func MergeCsv(files []string, final_file string) (error) {
   if Exists(final_file) == true {
		panic("File exists: " + final_file)
   }

  keys := make([]string, 0)

  for _, file := range files {
    var nfield,er = ParseCsv(file, true)
    if er != nil {
      fmt.Println("Error:", er)
      return er
    }
    for i:=0;i< len(nfield[0]);i++ {
	  inarr := false
	  name_field := strings.TrimSpace(strings.ToUpper(nfield[0][i]));
	  for n := 0; n < len(keys); n++ {
	  	if keys[n] == name_field {
			inarr = true
			break
		}
	  }
	  if inarr {
	  	continue
	  }
      keys = append(keys, name_field)
    }
  }
  
  writes := 0

  var fcontent = make([][]string, 0)
  fcontent = append(fcontent, keys)
  header_file := "/tmp/_merged_fields";

  // delete if exists
  if Exists(header_file) == true {
  	os.Remove(header_file);
  }
  WriteCsv(header_file, fcontent)

  paths := make([]string, 0)  
  paths = append(paths, header_file)

  //
  for _, path := range files {
    file, err := os.Open(path)
    if err != nil {
      fmt.Println("Error:", err)
      return err
    }

    reader := csv.NewReader(file)
  
    reader.Comma = ','
    reader.LazyQuotes = true
    reader.TrailingComma = true

    lineCount := 0
    lines := make([][]string, 0)
    nfield := make([]string, 0)
	
	hasher := md5.New()
    hasher.Write([]byte(path))
	uuid := hex.EncodeToString(hasher.Sum(nil))
	
	tmp_file := "/tmp/" + uuid + "_part"
    paths = append(paths, tmp_file)
  	// delete if exists
  	if Exists(tmp_file) == true {
  		os.Remove(tmp_file);
  	}
	
    for {
      record, err := reader.Read()
      if err == io.EOF {
        break
      } else if err != nil {
        return err;
      }

      lineCount++;
      if lineCount == 1 {
        for i:=0;i<len(record);i++ {
          record[i] = strings.TrimSpace(strings.ToUpper(record[i]))
        }
        nfield = record
      } else {
        values := make([]string, 0)
        // merge fields
        for i := 0; i < len(keys); i++ {
          value := ""
          pos := IndexOf(nfield, keys[i])
          if pos != -1 {
            value = record[pos]
          }  else {
            value = "N.D."
          }
          values = append(values, value)
        }

        lines = append(lines, values)
        values = nil

        if len(lines) > 1000 {
          writes++
          WriteCsv(tmp_file, lines)

          lines = nil
        }
      }
    }

    if len(lines) > 0 {
      writes++
      WriteCsv(tmp_file, lines)
      lines = nil
    }
  }

  // join files
  _files := strings.Join(paths, " ")    
  cmd := "/bin/cat " + _files + " > " + final_file
  out, err := exec.Command("bash", "-c", cmd).Output()
  if err != nil {
	fmt.Println(out)
    panic(err)
  }

  return nil
}

func Usage() {
	fmt.Println("Usage:\n")
	fmt.Println("\tmerger [final file] [files...]\n")
}

func main() {
  if len(os.Args) < 3 {
    Usage()
	os.Exit(2)
  }

  final_file := os.Args[1]

  files := os.Args[2:]

  // compilar archivos CSV
  MergeCsv(files, final_file)

}
