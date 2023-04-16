# aws-lambda-hsplog-extension
Lambda extension to push function logs directly to a HSP Logging service

## Build package and dependencies

To run this example, you will need to ensure that your build architecture matches that of the Lambda execution environment by compiling with `GOOS=linux` and `GOARCH=amd64` if you are not running in a Linux environment.

Building and saving package into a `bin/extensions` directory:
```bash
$ cd aws-lambda-hsplog-extension
$ GOOS=linux GOARCH=amd64 go build -o bin/extensions/aws-lambda-hsplog-extension main.go
$ chmod +x bin/extensions/aws-lambda-hsplog-extension
```

## Layer Setup Process
The extensions .zip file should contain a root directory called `extensions/`, where the extension executables are located. In this sample project we must include the `aws-lambda-hsplog-extension` binary.

Creating zip package for the extension:
```bash
$ cd bin
$ zip -r extension.zip extensions/
```

Publish a new layer using the `extension.zip` using below command. The output should provide you with a layer ARN. 

```bash
aws lambda publish-layer-version \
    --layer-name "aws-lambda-hsplog-extension" \
    --zip-file  "fileb://extension.zip"
```

Note the `LayerVersionArn` that is produced in the output. eg. 

```
LayerVersionArn: arn:aws:lambda:<region>:123456789012:layer:<layerName>:1
```

Add the newly created layer version to a Lambda function.

```bash
aws lambda update-function-configuration 
    --function-name <your function name> 
    --layers <layer arn>
```

## Function Invocation and Extension Execution

Configure the extension by setting below environment variables

* `HSDP_LOGGING_BASE_URI` - the base URI of HSP Logging service.Ex https://logingestor-dev.us-east.philips-healthsuite.com. 
* `PRODUCT_KEY` - Logging product key
* `SHARED_KEY` - Logging shared key
* `SECRET_KEY`-  Logging secret key

### Optional environment variables
* `ENABLE_VERBOSE_LOGGING` - setting this value to **true** logs the all of the logs from  **info level & above**
