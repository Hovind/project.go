# Network algorithm

* NO HEAD
	-> Connect to a node.
	-> Broadcast "LOCAL_IP connected to HEAD_IP".
* HAS HEAD
	- Receives "IP" connected to HEAD_IP.
		-> Connect to IP.
	- Does not receive Hamilton echo.
		-> Broadcast "HEAD_IP is mine, if you thought it was yours or have nil HEAD_IP, connect to LOCAL_IP instead".
