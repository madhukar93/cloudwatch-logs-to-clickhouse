
```
ShardIterator=$(awslocal kinesis get-shard-iterator --shard-id shardId-000000000000 --shard-iterator-type TRIM_HORIZON --stream-name api_logs_to_cloudwatch --region us-east-1 | jq -r ".ShardIterator")

{
    "ShardIterator": "AAAAAAAAAAFn8g3BAyOaLNfdJsTIoHOPt7YTAwRmTuqkobYsONtoeDCjBIhnRFu/nAh1BcI70mxFkipsqzyQqV+fBOoavcaVoArMggQaYJER+W1bx4iVDljNanfBOnW/EoNDORRz0gDpbRFbOUFR2Myc8M9bwC8sRyC/PNM9TPeKG22XAsS8POAYL/LarPof50jY2vQXktla2uFt2S2cCMLSLhSQNum5"
}

awslocal kinesis get-records --shard-iterator "$ShardIterator" --region us-east-1

{
		"records": [
        {
            "SequenceNumber": "49645468799910355314410530829901722170530862036047364098",
            "ApproximateArrivalTimestamp": 1697284916.341,
            "Data": "H4sIADSDKmUC/7VV207jMBD9FSvaR0LiXNrEb2UpCImb2rIrLUErx5mA1dzWdoAu4t/XadI0sFSCFdu3zBx7zpk54z4ZOUhJb2GxqsAgxuFkMfl5Np3PJ8dTY88oHwoQOmwPfjqclbfHoqwrnbHog7QymscJtXTYrESZ1EwfWqPmSgDNNcyxHdfCtoU96/rL6WQxnS9u/CAeO2HqpnY6ooGbhp4/pmMnHTt+4MR2oq+QdSyZ4JXiZXHEMwVCGuTaYBlny7uylmDSipu6kHGzrje9h0I1kCeDJ01Zz9W3KK5FKpprvngUjp3AC/HICYK9jXiNnC8mswWawa9aQ08SgnDsp67veSb2U2x6XuqZgRP6ph9S6kM8sqkN6JsmpKkR1Gkynvf6yt7OyiEeVm5aY2LbxN4CO8TGxHf3nbH7I1LvoRCpk/Oji0jNgAG/hwRB0wKCniKBUGQsoKCF1hMZRH+p9ZcTGXtd9o6L5JIKtTqnOXSYPnbQ46aF4mq1ucULA2Cub5u27yWmFzNqBjEOTJwAo0HKPN8JXx1t3NUebv2h+vxFBYI24/1KFdyWYrWBcQZXVaKDyd/YeR2/hMt1YCvsrGTLOdNqBS9bRAIZXUEyA1mVhdRkIvE8nJa/e1r+50+LlYWCx+2YGM2ymLLld8qVPCrFNK/Uam3m07KsGgVK1NCJS+uCNV3ozNfq6wzYd2AD2g52uJ09LIdcN/GU51ydFGcHLRI7QQ/YrPr2nl0bPzzS7v32zL+sf38fL+7LJSRHnaKJ6CRTURDNhbRcSC1NoFKZmAxfK7JpBHmTqz7f73wn/j1jfO2e0W73hJ/vnsbdSHb2JghlehcKthoSGu8i5NrufyBEl7y4RaLtJFINqTulKmJZd6VU+4kmDGKfa9OLgmbEC0MntN6gHeymPR7Snp4ffvStHpYJd5bBL57m2fTy4uP/CpE6rNt3iiBXW3DftlEuI3XAs0w/0C+T68zZegvRnP8GXcIJkN5E3dRH1CWuJCTbhPF88/wHeYvD+rgHAAA=",
            "PartitionKey": "/aws/lambda/log-producer",
            "EncryptionType": "NONE"
        }
    ],
    "NextShardIterator": "AAAAAAAAAAEAYPjVK/WntzXYRfZtREL1I5xkmOjlRCa3TMzstGZuSrlGDJP3z8bD+T+3emb2LQt3sfheqFUxBuGzu08yleRPgJBh8LrLu/T5/Vi49ekE5v7khn86Ssz5pReOUMJQl93NW/T2UYr2FBwBwTNyoUmq9AU4KgXkCDmzTNop+uUg0J6k+1BsxMjkdE2FRhwlgfH1niHdwlZBj/2wmrCplt5K",
    "MillisBehindLatest": 0
}


awslocal kinesis get-records --shard-iterator "$ShardIterator" --region us-east-1 | jq -r ".Records[0].Data" | base64 --decode | gunzip

```
