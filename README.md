# aconf
HOCON format reading/parsing library written in golang

## Motivation
Storing configuration information in human friendly format is a requirement common across projects of all sizes.  
The HOCON format, made popular by the Akka project, provides a specification which describes a configuration file format for precisely this purpose.  
There are many popular config file formats/API's in the world of golang - most popular being TOML, YAML, INI, etc. 
While these formats are popular, I believe none provide the flexibility provided by the HOCON format.  
The *aconf* library has been designed with the aim of attempting to allow Golang developers make full use of the HOCON configuration format using an extremely simple API.

## Getting the library
```
go get -u github.com/en-vee/aconf
```

## Features 
- Similar to the ```encoding/json``` library, one can Unmarshal/Decode a HOCON file into a go ```struct```
- Specify config properties as Units such as duration and size.
- Specify config properties as Arrays of primitives or arrays of objects

## API Usage
- Define the HOCON Configuration file. All property keys will need to start with a capital letter
```js
A {
    B = 10
    T = 25 seconds
    C = [1, 2, 3, 4]
}
```
- Create a HOCON parser by providing an ```io.Reader``` to read the file
```go
reader, err := os.Open("/path/to/configFilename.conf")
parser := &HoconParser{}
```
- Declare a go struct to match the configuration file format. Note that all the members of the go struct need to be exported/capitalized and the field names within the struct should exactly match the field names in the HOCON config file. At the time of writing, golang based Struct Tags are not supported.
```go
type ConfigFile struct {
    A struct {
        B int
        T time.Duration
        C []int
    }
}
```
- Call the Parse method to decode the Configuration file contents into the pointer to the struct
```go
var appConfig = &ConfigFile{}
if err := parser.Parse(reader, appConfig); err != nil {
			fmt.Printf("Error %v", err)
}
```
- If the Parse method returns without errors, the ```appConfig``` pointer in the example above will be populated with the values from the config file. For example, ```appConfig.B = 10``` or ```appConfig.T = 25s```
- Be sure to have a look at the parser_test.go file for various examples of Config file formats