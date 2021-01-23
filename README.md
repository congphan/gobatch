# gobatch
A simple Go library to handle data in batch

## Documentation
[![Go Reference](https://pkg.go.dev/badge/github.com/congphan/gobatch.svg)](https://pkg.go.dev/github.com/congphan/gobatch)

## Features
- Excute a function from given data by batch
- Stop execute next batch when receive cancel signal from given context
- Stop execute next batch when given context is timeout

## How to use
```
import "github.com/congphan/gobatch/processor"

employees := []employee{
    {
        Age:  50,
        Name: "name A1",
        Job:  "JOB 1",
    },
    {
        Age:  60,
        Name: "name A1",
        Job:  "JOB 2",
    },
    {
        Age:  70,
        Name: "ABC XYZ",
        Job:  "JOB 3",
    },
}

// simulate function to store list of employee to database
storeEmployees := func(s []employee) {
    fmt.Println(s)
}

p, _ := processor.New(2) // separate employees to batch with maximum 2 employees in a batch if enough
err := p.Execute(context.Background(), employees, func(batch processor.Batch) {
    data := batch.Data().([]employee)
    idx := batch.Index()

    fmt.Println("batch index: ", idx)
    storeEmployees(data)
})

```

## Examples
[![Go Reference](https://pkg.go.dev/badge/github.com/congphan/gobatch.svg)](https://pkg.go.dev/github.com/congphan/gobatch)

## Benchmark Results
| source size| batch size | ns/op | B/op  | allocs/op|
|:-----------|:-----------|------:|------:|---------:|
|1000        |100         |8475   |1304   |22        |
|10000       |100         |86139  |12824  |202       |
|100000      |100         |821887 |128024 |2002      |
|1000000     |100         |8087827|1280063|20002     |

## Incoming features
- Provide interface for data source
- Batch job

## License
gobatch is provided under the MIT licence. See LICENSE for more details.
