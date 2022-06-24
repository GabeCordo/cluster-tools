### Net Package

#### Request
The request data structure and accompanying functions are used to format and send data in the form of HTTP JSON compatible requests. 
Most requirements for JSON key naming, ECDSA authentication signing, and networking related functionality have been implemented within the associative struct functions.

```[INFO] If you need to send data to a Node, do NOT use the http struct. You will find yourself re-implementing most features provided by the Request abstraction.```

##### Send(method, url string)
##### Sign()