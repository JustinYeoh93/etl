# lambda
Lambda is the component that contains the code to execute in the event of a trigger from SQS.

Most of our logical code is written in the `extractor` directory, whereas `lambda` only contains the infrastructure setup required to post our code to the serverless environment.