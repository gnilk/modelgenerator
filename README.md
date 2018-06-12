# modelgenerator
Golang POGO domain model generator with DB CRUD support

Specify your domain model in XML, run it through the model generator and it will spit out a go definition plus a DB Crud layer (using MYSQL).

Example:
<?xml version="1.0" encoding="UTF-8"?>
<doc namespace="resource">

    <imports>
        <package no_persistence="true">uuid github.com/satori/go.uuid</package>
        <package>time</package>
    </imports>

    <define type="class" name="Resource">
        <guid name="ResourceID" />
        <guid name="UserID" />
        <guid name="EntityID" />        
        <string name="Filename" />
        <string name="Path" />
        <string name="MimeType" />
        <bool name="IsEntityResource" />
        <bool name="External" />
        <time name="CreateDate" />
        <time name="LastUpdateDate" />
        <list type="byte" name="Data" />
    </define>
</doc>

You can specify additional imports in the 'imports' section. These will be placed on top of your GO datamodel file.

An XML document can hold many classes - each document will generate on GO file (there is an experimental option to split per class but - but I never use it).

Use the tool like:
modelgenerator v2 - XML Data Model to GO structure converter
Usage: modelgenerator [-sv] [-p <class>] [-f <num>] [-o <file/dir>] <inputfile>
Options
  -f : From Version, generates any class/field matching >= specified version (0 means as virgin)
  -p : Generate persistence (use optional 'class' to specifiy which class for persistence, or '-' for all - default)
  -d : Generate drop statements before create (default = false)
  -s : split each type in separate file
  -c : generate convertes (to/from XML/JSON)
  -o : specify output model go file or dir (if split in multiplefiles is true), default is stdout
  -O : specify output database go file or dir (if split in multiplefiles is true), default is 'db.go'
  -v : increase verbose output (default 0 - none)
  -h : this page
inputfile : XML Data Model definition file

Example invocation:
modelgenerator -v -p -c file.xml -o file.go

It is advisable to run GOIMPORTS on the generated file - that way you can have a common set of imports in your domain and GOIMPORTS will strip what's not used.
The tool support type-mapping from the XML definition to GO and MYSQL types.
Like:
    <dbtypemappings>
        <map from="guid" to="varchar(36)" />
        <map from="string" to="varchar(%d)" fieldsize="128"/>
        <map from="time" to="datetime" />
        <map from="EmploymentType" to="int(11)" />
    </dbtypemappings>

    <gotypemappings>
        <map from="guid" to="uuid.UUID" />
        <map from="time" to="time.Time" />
    </gotypemappings>

This allows the 'type' declaration to be transformed properly when generating the GO code.
The tool allows for language extensions but so far only GO is supported.

Following ROOT tags are supported:
* include - allow include of other documents to this document (this is a simple 'add' from the included document)
* dbtypemappings - type mapping control for DB CRUD generator
* gotypemappings - type mapping controil for GO language
* dbcontrol - specification of common attributes for the DB layer (user, schema, etc..)
* imports - GO language imports
* define - definintion of a data type (enum or class)







