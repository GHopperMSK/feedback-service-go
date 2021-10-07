## Deployment

### deploy env
$ docker-compose up

### build binary
$ go build -o bin ./cmd/*

### connect to DB
$ docker exec -it db sh
$ mysql -u db_user feedback_service -p

### create kafka event
$ echo "{\"action\":\"create-action\",\"version\":\"v0.1\",\"payload\":{\"sender_uuid\":\"807a51d6-a81b-4b66-9596-5b17ea26b136\",\"sender_name\":\"sender#1\",\"sender_avatar\":\"sender#1 avatar\",\"receiver_uuid\":\"807a51d6-a81b-4b66-9596-5b17ea26b137\",\"receiver_name\":\"receiver#1\",\"receiver_avatar\":\"receiver#1 avatar\",\"offer_hash\":\"ksO3jso7aDi\",\"offer_authorized\":true,\"offer_owner_uuid\":\"807a51d6-a81b-4b66-9596-5b17ea26b138\",\"offer_type\":\"SELL\",\"offer_payment_method\":\"PayPal\",\"offer_payment_method_slug\":\"paypal_slug\",\"offer_currency_code\":\"RUB\",\"trade_hash\":\"isO9AlIU8s2\",\"trade_fiat_amount_requested_in_usd\":\"320.12\",\"trade_status\":\"RELEASED\",\"message\":\"message1\",\"feedback_type\":\"POSITIVE\",\"created_at\":\"2014-11-12 11:45:26.37\"}}" | kcat -P -b localhost:29092 -t test -p 0

$ echo "{\"action\":\"create-action\",\"version\":\"v0.1\",\"payload\":{\"sender_uuid\":\"807a51d6-a81b-4b66-9596-5b17ea26b136\",\"sender_name\":\"sender#1\",\"sender_avatar\":\"sender#1 avatar\",\"receiver_uuid\":\"807a51d6-a81b-4b66-9596-5b17ea26b139\",\"receiver_name\":\"receiver#2\",\"receiver_avatar\":\"receiver#2 avatar\",\"offer_hash\":\"A3O3jso7aUi\",\"offer_authorized\":false,\"offer_owner_uuid\":\"807a51d6-a81b-4b66-9596-5b17ea26b136\",\"offer_type\":\"BUY\",\"offer_payment_method\":\"SEPA\",\"offer_payment_method_slug\":\"sepa_slug\",\"offer_currency_code\":\"EUR\",\"trade_hash\":\"tsO9Al83k8s\",\"trade_fiat_amount_requested_in_usd\":\"20.32\",\"trade_status\":\"RELEASED\",\"message\":\"message2\",\"feedback_type\":\"POSITIVE\",\"created_at\":\"2016-11-12 11:45:26.37\"}}" | kcat -P -b localhost:29092 -t test -p 0

$ echo "{\"action\":\"update-action\",\"version\":\"v0.1\",\"payload\":{\"feedback_id\":1,\"message\":\"message1 NEW\",\"feedback_type\":\"POSITIVE\"}}" | kcat -P -b localhost:29092 -t test -p 0

$ echo "{\"action\":\"delete-action\",\"version\":\"v0.1\",\"payload\":{\"feedback_id\":2}}" | kcat -P -b localhost:29092 -t test -p 0

$ echo "{\"action\":\"delete-offer-action\",\"version\":\"v0.1\",\"payload\":{\"offer_hash\":"ksO3jso7aDi", \"deleted_at\":\"2014-11-12 11:45:26.37\"}}" | kcat -P -b localhost:29092 -t test -p 0

$ echo "{\"action\":\"change-trade-status-action\",\"version\":\"v0.1\",\"payload\":{\"trade_hash\":"ksO3jso7aDi", \"status\":\"DISPUTED\"}}" | kcat -P -b localhost:29092 -t test -p 0

## Run tests

$ docker-compose -f docker-compose.yml -f docker-compose.test.yml up


