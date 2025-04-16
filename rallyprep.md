# Rally documentation guidelines
This is aimed at Rallymasters preparing documentation to be used in an event they're designing.

## What is 'data'
Preparation of the rally will involve the production of several discrete clumps of 'data'. The details of each bonus, the text of the instructions, the rally rules, the various images, etc. Except for the images, the appearance of data items is irrelevant as is the way they're laid out on a page. In addition to the rally book itself, the data will be used to build the database used for scoring purposes so it's unlikely that a single word-processed document will be adequate.

In general, each item of data should be maintained separately using suitable software. A word-processor is fine for the text but keep any styling simple. Images should have simple consistent names (lettercase matters) and formats (jpg, png). Keep images in a separate folder which can then be supplied as, for example, a zip archive. Spreadsheets can be used to hold bonus details, combo details, etc.

## Presentation guidance
Use a word-processor to mock up a rally book so that it looks like the desired finished product. Convert the document to PDF format when sending to anyone else.

## Scoring strategy
The most common approach to scoring is that each bonus is worth a fixed number of points and scores are built by visiting several bonuses and accumulating the points. Combos offer a second level of scoring along the same lines. How scores are built for individual rallies is limited only by the Rallymaster's imagination *and*, most importantly, the quality of the explanation of the methods provided to entrants and rally teams.

## Bonus specification
### Bonus codes
Each bonus needs to be uniquely identified by a short code. Early rallies all used simple number sequences for this purpose but the codes need not be numeric or sequential. The 2023 IBR codes comprised the initials of the state and city in which each bonus was located - FLCK = Florida, Cedar Keys; GAAT = Georgia, Atlanta and so on. If letters are used, they must be uppercase.

If the codes will be a simple numeric sequence, take into consideration that they will be treated by computers as text, not numbers. The normal order of a sequence starting at 1 would be 1, 10, 11, ...  18, 19, 2, 20, ... which might confuse some people. The order can be "fixed" by including leading '0's as necessary so 01, 02, 03, ... 09, 10, ... but that also confuses some people who don't understand that '1' is not the same as '01'.

A 90 bonus sequence could use the range '10' ... '99' which avoids both of those issues. If more than 90, start at 100 or 101.

Avoid codes where the difference between '0' (zero) and 'O' (oh) might not be obvious.


### Description
Bonus descriptions appear in the rally book, on scorecards and claim records. They should be kept short and with a consistent layout (state, city; author, title; county, building type, etc). The descriptions should not contain 'data' relied on elsewhere, they are just arbitrary text.

### Points
Bonus points are a simple integer value (which can be negative!). It is possible to specify a variable bonus where, for example, the actual value is entered manually by a member of the rally team. Variable bonuses are used with instructions such as "read the clock". Thanks to Phil Weston it is also possible to use a bonus as being a multiplier applied to an earlier claim rather than having its own points value.

### Flags
Each bonus can be flagged with one or more of:-
- A - Alert, read notes carefully
- B - Bike must be in photo
- D - Daylight only
- F - Face must be in photo
- N - Nighttime only
- R - Restricted hours/access
- T - Ticket or receipt required

### Scoring notes
This is the text of instructions for claiming the bonus. The basic "park and take the same photo" is unnecessary, these notes should be confined to, for example, "photo the south side, not the north". Scoring judges will use this information when deciding claims.

### Bonus information
This is called "waffle" in the scoring system database. This is whatever you want to say about the bonus not directly concerned with how to score it.

### Categories
If part of your scoring strategy involves the use of categories (counties, countries, activity, building type, etc) you need to supply a list of the category values and you'll need to mark each bonus as belonging to whichever categories are appropriate. In the database, individual category values are identified by a unique integer (1=Surrey, 2=Hampshire, ...) but you can use short codes or full titles, whichever is easier.

### Coords
If coordinates are to be printed in the rally book or used to create a GPX file, they can be supplied in any recognisable format.

### Image
The name (lettercase matters) of the rally book image of this bonus.

### Rest minutes
If a bonus counts towards a rest bonus, this value specifies the amount of time scored.

### Question / answer
If a question and answer component features in the rally, these fields hold the individual questions and answers.

## Standard texts
Standard texts exist for the rally rules, how to make EBC claims, how to take photos, etc. In almost all cases these standard texts should be used in preference to any custom wording.

Standard texts can be included by reference to their title (IBAUK Rally Rules 2023) or by including the actual text. If the latter, include a note of any changes from the standard or assert that there are no changes.