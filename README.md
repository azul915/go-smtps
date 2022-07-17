# docker-postfix-tls #



## build ##

    $ docker-compose build


## run ##

    $ docker-compose up


## mail dev の確認方法 ##

ブラウザで http://localhost:1080/ にアクセスする。

## command ホストで、メール送信 ##

### command サービスに入り、mailxコマンドでメールを送信する。
    
```sh
docker-compose exec command sh
```
    
```sh
echo 'mail body' | mailx \
   -v \
   -S smtp-use-starttls \
   -S smtp-auth=login \
   -S ssl-verify=ignore \
   -S smtp="postfix:587" \
   -S smtp-auth-user="smtpuser" \
   -S smtp-auth-password="password" \
   -S from="hoge@example.com" \
   -r "hoge@example.com" \
   -s "TLS test mail" \
   "foo@example.com"
```

### 実行結果
```sh
% docker-compose exec command sh
/ # echo 'mail body' | mailx \
> -v \
> -S smtp-use-starttls \
> -S smtp-auth=login \
> -S ssl-verify=ignore \
> -S smtp="postfix:587" \
> -S smtp-auth-user="smtpuser" \
> -S smtp-auth-password="password" \
> -S from="hoge@example.com" \
> -r "hoge@example.com" \
> -s "TLS test mail" \
> "foo@example.com"
Resolving host postfix . . . done.
Connecting to 192.168.0.3:587 . . . connected.
220 postfix.example.com ESMTP Postfix (Ubuntu)
>>> EHLO 60aad46491db
250-postfix.example.com
250-PIPELINING
250-SIZE 10240000
250-VRFY
250-ETRN
250-STARTTLS
250-AUTH PLAIN LOGIN CRAM-MD5 DIGEST-MD5 NTLM
250-AUTH=PLAIN LOGIN CRAM-MD5 DIGEST-MD5 NTLM
250-ENHANCEDSTATUSCODES
250-8BITMIME
250 DSN
>>> STARTTLS
220 2.0.0 Ready to start TLS
>>> EHLO 60aad46491db
250-postfix.example.com
250-PIPELINING
250-SIZE 10240000
250-VRFY
250-ETRN
250-AUTH PLAIN LOGIN CRAM-MD5 DIGEST-MD5 NTLM
250-AUTH=PLAIN LOGIN CRAM-MD5 DIGEST-MD5 NTLM
250-ENHANCEDSTATUSCODES
250-8BITMIME
250 DSN
>>> AUTH LOGIN
334 VXNlcm5hbWU6
>>> c210cHVzZXI=
334 UGFzc3dvcmQ6
>>> cGFzc3dvcmQ=
235 2.7.0 Authentication successful
>>> MAIL FROM:<hoge@example.com>
250 2.1.0 Ok
>>> RCPT TO:<foo@example.com>
250 2.1.5 Ok
>>> DATA
354 End data with <CR><LF>.<CR><LF>
>>> .
250 2.0.0 Ok: queued as 6EC3A423EBE
>>> QUIT
221 2.0.0 Bye
/ # exit
```
