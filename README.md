# longChat
基于websocket的聊天系统

##Features
1.扩展伸缩方便，树型架构决定了系统是低耦合的，节点可以随意热拔插

2.高可用，所有节点上每一台服务器运行的都是同一份代码，一台down了时可以使用任意一台顶替

3.基于社交网络的聚类(Clustering)，将联系频繁的用户划分到同一节点服务器，减少根节点服务器压力

##Architecture  

![](http://o8up60qgx.bkt.clouddn.com/%E6%9E%84%E6%9E%B6.png)  

