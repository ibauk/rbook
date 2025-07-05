# rbook - A Rally Book generator

I prepare rally books and, optionally, GPX files for IBA rallies using data held in a [ScoreMaster](https://github.com/ibauk/sm3) database and templates coded in HTML & CSS and images stored on disk. My output is a single HTML document ready for printing to PDF format.

Each run is controlled by a standard YAML configuration file identified by the *-cfg* commandline variable, default "std.yml". The parameters are:-

## title
The title of the rally, used as the title of the output document. This is overwritten using the RallyTitle field from the ScoreMaster database.

## description
Additional description of the edition. Perhaps "A4 portrait, 2 column photos"

## projectFolder
The filepath to the folder containing templates for this project.

## outputFolder
The filepath to the folder that will hold the output files.

## rallybookFile
The name of the file containing the generated HTML. This can be overriden using the *-book* commandline variable.

## database
The filepath to the ScoreMaster database used with this project. This can be overriden using the *-db* commandline variable.


## imageFolder
URL, relative to outputfolder, to folder containing images. This would normally point to the **sm/images** folder of a ScoreMaster installation with bonus images held in **sm/images/bonuses**. A typical bonus image inclusion in a template might be `{{.ImageFolder}}/bonuses/01.png`.


## landscape
true/false. The default is false, portrait mode. This chooses between CSS files.


## sections
This holds a list of templates to be processed in sequence. A template can be either a static template or a 'stream' which is applied either to a selection of bonuses or a selection of combos. Stream templates are identified in this list by the prefix `stream.`. The template names listed here will have `.html` appended to identify the file on disk.

## streams
This holds a list of stream specifications. Each specification includes the following fields:-

### streamid
The *template* name of this stream. The name included as *stream.streamid* above referring to the file *streamid.html* on disk.

### type
What type of template this is. One of `static`, `bonus`, or `combo`.

### whereSQL
The SQL string to follow the WHERE in the SELECT string for bonuses or combos.

### orderByField
The SQL string to follow the ORDER BY in the SELECT string for bonuses or combos.

### colsPerRow
The number of bonuses or combos to be output to a single line across the page.

### rowsPerPage
The number of rows to be output per printed page.

### emitGPX
true or false - rows from this stream should be included in any GPX file.

### generateGPX
#### outputFile
The path of the output GPX file relative to the *outputFolder*. Can be overridden with *-gpx*  option. If left blank, no GPX file is created.

#### link2map
The url of a mapping service. The link will have latitude and longitude appended automatically. 
eg: https://www.google.co.uk/maps/search/

#### symbol
The quoted name of a recognised symbol to be used. eg: "Circle, Green"

#### bonusidOnly
If false, the waypoint name will be *bonusid* - *briefdesc*
---

## Sample config 

```
{


    title: 2022 Invictus Tour

    description: A4 portrait, 2 column photos

    projectFolder: projects\sample\

    outputFolder: \sm-installs\testmodel

    database: \sm-installs\testmodel\sm\scoremaster.db


    imageFolder: sm/images/


    landscape: false


    sections: [ frontpage, header, walkhdr, stream.bonuses, footer]


    streams:
        - { 
            streamid:     bonuses, 
            type:         bonus,
            whereSQL:     BonusID NOT LIKE '%-' AND BonusID <='14',
            orderByField: BonusID,
            colsPerRow:   2
            rowsPerPage:  3
            emitGPX:      true
          }

    generateGPX:
        outputFile: it22.gpx
        symbol: "Circle, Green"



}
```

---
## Template files
These consist of ordinary HTML/CSS files with data field inclusion specified by the characters `{{}}`. 

For static templates the possible inclusions are the *CamelCase* versions of the YAML keys above (ImageFolder, ProjectFolder, etc). 

For bonus streams: Any of the fields in the bonus record + ImageFolder, NewLine flag, StreamID and the scoring flags (AlertT, AlertR, AlertF, AlertB, AlertD, AlertA).



