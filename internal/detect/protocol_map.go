package detect

var ProtocolForService map[string]string

func init() {
	output := map[string]string{}
	for protocol, services := range protocolUsage {
		for _, service := range services {
			output[service] = protocol
		}
	}
	ProtocolForService = output
}

// ProtocolUsage provides a correspondence from AWS API protocols to AWS services
var protocolUsage = map[string][]string{
	"ec2query": []string{"ec2"},
	"restxml":  []string{"cloudfront", "route53", "s3"},
	"query": []string{
		"autoscaling",
		"cloudformation",
		"cloudsearch",
		"cloudwatch",
		"elasticache",
		"elasticbeanstalk",
		"elb",
		"iam",
		"rds",
		"redshift",
		"ses",
		"simpledb",
		"sns",
		"sqs",
		"sts",
	},
	"restjson": []string{
		"apigateway",
		"cloudsearchdomain",
		"cognitosync",
		"efs",
		"elasticsearchservice",
		"elastictranscoder",
		"glacier",
		"iot",
		"iotdataplane",
		"lambda",
		"mobileanalytics",
	},
	"jsonrpc": []string{
		"cloudhsm",
		"cloudtrail",
		"cloudwatchlogs",
		"codecommit",
		"codedeploy",
		"codepipeline",
		"cognitoidentity",
		"configservice",
		"datapipeline",
		"devicefarm",
		"directconnect",
		"directoryservice",
		"dynamodb",
		"dynamodbstreams",
		"ecs",
		"emr",
		"firehose",
		"inspector",
		"kinesis",
		"kms",
		"machinelearning",
		"marketplacecommerceanalytics",
		"opsworks",
		"route53domains",
		"ssm",
		"storagegateway",
		"support",
		"swf",
		"waf",
		"workspaces",
	},
}
