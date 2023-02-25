# ETL
This is an ETL pipeline that scrapes e-commerce API endpoints for specific information and then store them in our data structure.

# Components
Below are the components of the ETL pipeline
- brain
- extractor

## Brain
The brain is the controller of the ETL. It decides what is to be scraped and communicate it to a centralized queue for processing.

The workflow of the `brain` is 
- get a list of EC info from a source
- get latest extraction status
- decide what needs to be processed
- appends to queue on what needs to be processed
- repeat at next cycle

## Extractor
Extractors a short lived process tasked with extracting information based on given configurations.

The workflow of the `extractor` is
- spins up for a given trigger
- parse given information about a extraction source
- extract info from the source
- updates the database with the new data
- shutdown