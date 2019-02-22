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
- Create a HOCON parser
- Declare a go struct to match the configuration file format. Note that all the members of the go struct need to be exported/capitalized and the field names within the struct should exactly match the field names in the HOCON config file
- Call the Parse method