# HTML Link Parser
- Random HTML Link Parser written in Go
> [!CAUTION]
> This is a personal project created for my own learning purposes. **This project is not intended for use with production data or in a production environment.**

## Overview
- Access a HTML page and all relative paths of the site via API calls concurrently, ensuring each path is accessible.
- Database: Postgresql. No bulk insertion/batch update/link repository.

## Features
- **Access the HTML page:**
  - The initial HTML page is accessed via an API call.
- **Extract anchor elements:**
  - Extract all outer anchor elements with an href attribute from the HTML page.
  - Extract and join all inner text elements within these anchor elements using a ", " delimiter.
- **Save the links:**
  - Save the extracted links and text into a MySQL database.
- **Retrieve and process the relative paths**
  - Recursively retrieve all the relative paths from database.
  - For each relative path, access the page, extract anchor links, and save them.
- **Note**:
  - Inner nested anchor links are omitted. Nesting anchor elements is not standard HTML practice.
  - Only links that are contained within anchor elements and are properly linked will be extracted. For example, if a page like "about-us" is not referenced or connected within other pages, it may not be included in the extraction.
  - Other links that are not relative paths are not being utilized at the moment. However, they can be used for future reference and analysis.

## Credits
Special thanks to https://github.com/gophercises/link for this practise idea.

## Other Ideas/Enhancements
- Explore batch inserts/updates for database queries instead of inserting/updating the link individually.
- Host my project and database in Kubernetes

## Contributions
This is a personal project for my own learning purposes, so I am not accepting changes or pull requests. However, I welcome new ideas and suggestions for improving the project.

## What I Have Learnt
### 29/08/2024
- Introduced concurrency to my program such that I can process more than 1 relative path concurrently.
- Intoduced a map and RWLock to optimize my program such that links that are saved in database will not need to be extracted again.
- Changed the process to remove whitespace -> Trim.Space() function, then Regex.
- Added mocks and test cases for database queries.

### 28/08/2024
**Whitespace**
- Trim.Space() function is applicable for leading and trailing whitespace.
- Used Regex to remove other whitespace such as \t, \n.
- However, the time taken for Regex to process and remove whitespace is time-consuming. My program takes 1s to process 89 links.

**Nested Anchor Elements**
- ex5.html (test case) originally in this project will be removed for the following reasons:
  - This file contains multiple levels of nested anchor (<a>) elements. When it is loaded in a browser, some of the anchor element links appear duplicated.
  - Inner nested anchor links are not allowed (https://www.w3.org/TR/PR-html40-971107/struct/links.html#h-12.2.2)
  - Anchor elements must not have interactive content (https://html.spec.whatwg.org/#the-a-element)
  - Conducted a test on nested anchor elements. If an anchor element (<a>) is nested directly inside another anchor element, modern browsers will generally treat them as separate, sibling elements, rather than as nested elements.


### 22/08/2024
- (Refer to update on 28/08/2024) ~~html.Parse function may have issues if there are other elements nested between two anchor tags. Refer to ex5.html (test case) generated by ChatGPT. To investigate and go through more test cases in the future.~~
- Adding Test Cases.

### 21/08/2024
**Naming Conventions for functions/methods**
- Keep it simple, avoid excessive reptition.


