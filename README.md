# 前言
在网上看了一堆go和智能合约交互的教程，大部分都是抄袭的，一抄二，二抄三。加上现在网络环境不好经常被墙，搞半天搞不完。本试验环境win10，例子参考官方文档。
remix + 测试网 + abigen + golandIDE

# 第一步写合约
```solidity
// SPDX-License-Identifier: GPL-3.0

pragma solidity >0.7.0 < 0.9.0;
/**
* @title Storage
* @dev store or retrieve variable value
*/

contract Storage {

	uint256 value;

	function store(uint256 number) public{
		value = number;
	}

	function retrieve() public view returns (uint256){
		return value;
	}
```
# 第二步 编译合约加部署
![在这里插入图片描述](https://img-blog.csdnimg.cn/d135135d2ea24b1e9152953dfa94a57c.png)
选择injected Provider 唤起小狐狸部署
![在这里插入图片描述](https://img-blog.csdnimg.cn/756e97840ecc4bf8b8b7137ac6f730a0.png)
# 第三步 安装go-ethereum
这里网上大部分会让你在github下载，然后让你go build 或者是其他，但是我这边网络就算翻了墙配置好代理也会超时。这里原来实际上是abigen是geth的一个开发工具，go-ethereum就是go语言实现的geth而已，里面会有很多个不同工具，你在仓库的readme就可以看到会有make geth 或者make all
但是我们直接点，直接下安装包安装，更加省心。
地址：[https://geth.ethereum.org/downloads/](https://geth.ethereum.org/downloads/)
选windows
![在这里插入图片描述](https://img-blog.csdnimg.cn/e96016c100f14222b584f762656c3f26.png)
安装选![在这里插入图片描述](https://img-blog.csdnimg.cn/1bf122ede94b4725a29adac1ccdc2f7f.png)
然后会出现一个PATH的报错（可能也没有）这个时候去电脑环境变量把安装地址加入进PATH即可
![在这里插入图片描述](https://img-blog.csdnimg.cn/4787f900830343089f7e2f95da642771.png)
这样就安装好了
# 第三步通过abi生成go文件
abi获取方法，在remix那复制
![在这里插入图片描述](https://img-blog.csdnimg.cn/66519e7bdb3f4cd2b6eaa1a25f44f3f7.png)
然后在命令行
```cmd
abigen --abi Storage.abi --pkg main --type Storage --out Storage.go --bin Storage.bin
```
这个时候就会生成一个go文件，这里面的参数自己看看就知道啥意思了不解释了。
然后用goland打开这个文件夹
```cmd
go mod init test
go mod tidy
```
初始化项目
然后再新建一个go文件来与合约交互
```go
package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Create an IPC based RPC connection to a remote node
	// NOTE update the path to the ipc file! 
	conn, err := ethclient.Dial("/home/go-ethereum/goerli/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	// Instantiate the contract and display its name
	// NOTE update the deployment address!
	store, err := NewStorage(common.HexToAddress("0x21e6fc92f93c8a1bb41e2be64b4e1f88a54d3576"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate Storage contract: %v", err)
	}

```
这里的ethclient.Dial()里面实际上要填一个客户端地址，这里我们去https://infura.io/zh注册申请一个
![在这里插入图片描述](https://img-blog.csdnimg.cn/f6b180d940134656bf54baae08c24c20.png)
选好对应的测试网络
![在这里插入图片描述](https://img-blog.csdnimg.cn/8f9c80cf560046b2af9df180d075b854.png)NewStorage(common.HexToAddress("0x21e6fc92f93c8a1bb41e2be64b4e1f88a54d3576"), conn)
填入合约地址，这里就是new一个调用合约的实例，很多方法都封装好，更加容易调用。
然后就可以访问链上合约的数据了。
# 和合约进行交易
交易和普通查询不同，需要私钥。并且对代码做点改动
代码
```go
package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

const key = `json object from keystore`

func main() {
	PrivateKey, _ := crypto.HexToECDSA("你的私钥")

	// Create an IPC based RPC connection to a remote node and instantiate a contract binding
	conn, err := ethclient.Dial("你的节点地址")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	store, err := NewStorage(common.HexToAddress("你的合约地址"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Storage contract: %v", err)
	}
	// Create an authorized transactor and call the store function
	nonce, _ := conn.NonceAt(context.Background(), common.HexToAddress("你私钥对应的账户地址"), nil)
	gasPrice, _ := conn.SuggestGasPrice(context.Background())
	//用哪条链，就用那个id
	auth, err := bind.NewKeyedTransactorWithChainID(PrivateKey, big.NewInt(5))
	auth.GasLimit = uint64(300000)
	auth.Nonce = new(big.Int).SetUint64(nonce)
	auth.GasPrice = gasPrice
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	// Call the store() function
	tx, err := store.Store(auth, big.NewInt(420))
	if err != nil {
		log.Fatalf("Failed to update value: %v", err)
	}
	fmt.Printf("Update pending: 0x%x\n", tx.Hash())

}

```
核心步骤，获取nonce，获取gasprice，绑定，发交易

还有一种方式是直接通过读abi文件就能发交易，那种的话大概差不多只是写代码没那么简洁。

就这样吧，希望对大家有帮助

# 参考资料
[https://geth.ethereum.org/docs/dapp/native-bindings](https://geth.ethereum.org/docs/dapp/native-bindings)
[https://medium.com/nerd-for-tech/smart-contract-with-golang-d208c92848a9](https://medium.com/nerd-for-tech/smart-contract-with-golang-d208c92848a9)
