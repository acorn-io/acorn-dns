# Ensure that the AWS credentials are set before proceeding
if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    echo "AWS_ACCESS_KEY_ID is not set, please set it and try again"
    exit 1
fi
if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
    echo "AWS_SECRET_ACCESS_KEY is not set, please set it and try again"
    exit 1
fi
if [ -z "$AWS_SESSION_TOKEN" ]; then
    echo "AWS_SESSION_TOKEN is not set, please set it and try again"
    exit 1
fi

# Replace the current aws-creds secret if it exists
acorn secrets aws-creds > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "aws-creds secret already exists, deleting it..."
    acorn secret rm aws-creds > /dev/null 2>&1
fi

acorn secret create aws-creds \
	--data access-key=$AWS_ACCESS_KEY_ID \
	--data secret-key=$AWS_SECRET_ACCESS_KEY \
	--data session-token=$AWS_SESSION_TOKEN \
    > /dev/null

# Ensure that the ROUTE53_ZONE_ID is set before proceeding
if [ -z "$ROUTE53_ZONE_ID" ]; then
    echo "ROUTE53_ZONE_ID is not set, please set it and try again"
    exit 1
fi

# Hand-off start to acorn and link the aws-creds secret
acorn dev -p 4315:4315 -s aws-creds:aws-creds . --route53-zone-id $ROUTE53_ZONE_ID