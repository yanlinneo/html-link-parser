![html-link-parser](https://github.com/user-attachments/assets/ff13d720-c37d-4683-9253-1afc0060e93c)
# HTML Link Parser
- Random HTML Link Parser written in Go

## Disclaimer
> [!CAUTION]
> **This project is not intended for use with production data or in a production environment.**
> 
> This is a personal project created for my own learning purposes.
> 
> I am not responsible for any issues that may arise from testing, including but not limited to:
> 
> - **Access Blocking:** Your IP or network might get blocked by servers if too many requests are made, or if requests are considered suspicious.
>   
> - **Security Risks:** Avoid using URLs from untrusted sources to protect against potential security threats.
>   
> - **Sensitive Data Exposure:** Be cautious with URLs that might contain sensitive or private information.

## Overview
- CLI app.
- Access a HTML page and all relative paths of the site via API calls concurrently, ensuring each path is accessible.
- Results are then written to a CSV file and saved in the root directory.
<img width="793" alt="CLI_Example" src="https://github.com/user-attachments/assets/e7c7e76e-8a5b-4c55-93ae-200e0b6b7bd7">


## Installation
1. Ensure you have PostgreSQL and Go installed.

2. Create the schema and table in PostgreSQL database.
```ruby
-- DROP SCHEMA html_link_parser;

CREATE SCHEMA html_link_parser AUTHORIZATION pg_database_owner;

-- DROP SEQUENCE html_link_parser.link_id_seq;

CREATE SEQUENCE html_link_parser.link_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 2147483647
	START 1
	CACHE 1
	NO CYCLE;-- html_link_parser.link definition

-- Drop table

-- DROP TABLE html_link_parser.link;

CREATE TABLE html_link_parser.link (
	id serial4 NOT NULL,
	url varchar NOT NULL,
	description varchar NULL,
	source_url varchar NULL,
	base_url varchar(100) NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	status_code int4 NULL,
	status_message varchar NULL,
	CONSTRAINT link_pkey PRIMARY KEY (id),
	CONSTRAINT url_base_url_uk UNIQUE (url, base_url)
);
```
3. In terminal, navigate to your preferred directory and install HTML Link Parser CLI app.
```
git clone https://github.com/yanlinneo/html-link-parser
```

4. To connect the app to database, we need to export the user and password in CLI app. In terminal, run the below commands.
```
export DBUSER=<your PostgreSQL database username>
export DBPASS=<your PostgreSQL database password>
```
5. Run the app with an URL flag. Ensure the URL has a scheme (e.g. https://) and host (e.g. github.com).
```
go run . -url=https://this-is-an-example-url.com
```
6. When the program completes extracting and processing the relative paths, the results will be written to a CSV file and saved within the root directory.
<img width="630" alt="CSV_Example" src="https://github.com/user-attachments/assets/dc811a85-26fa-48cc-8d41-df642090b46b">

## Features
- **Access the HTML page:**
  - The initial HTML page is accessed via an API call.
- **Extract anchor elements:**
  - Extract all outer anchor elements with an href attribute from the HTML page.
  - Extract and join all inner text elements within these anchor elements using a ", " delimiter.
- **Save the links:**
  - Save the extracted links and text into a database.
- **Retrieve and process the relative paths**
  - Recursively retrieve all the relative paths from database.
  - For each relative path, access the page, extract anchor links, and save them.
- **Note**:
  - Inner nested anchor links are omitted. Nesting anchor elements is not standard HTML practice.
  - Only links that are contained within anchor elements and are properly linked will be extracted. For example, if a page like "about-us" is not referenced or connected within other pages, it may not be included in the extraction.
  - Other links that are not relative paths are not being utilized at the moment. However, they can be used for future reference and analysis.
  - Fragment identifiers are excluded.

## Credits
Special thanks to https://github.com/gophercises/link for this practise idea.

## Other Ideas/Enhancements
- Host my project and database in Kubernetes.

## Contributions
This is a personal project for my own learning purposes, so I am not accepting changes or pull requests. However, I welcome new ideas and suggestions for improving the project.

