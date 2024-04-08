# Ollama Chat Bot
> 本地大模型的聊天机器人

## 介绍
本项目展示一个简单易用的本地大模型聊天机器人，基于[Ollama](https://github.com/ollama/ollama)开发

## 为什么没有直接用Ollama服务
Ollama已经做的足够好了，但对小白来说还不够简单。 先要Ollama以服务形式启动，然后再启动应用层调用Ollama的接口。
这对于一些小白来说还是有一定的门槛。所以本项目是把Ollama以库形式直接嵌入应用层，一键启动，无需其他关联服务。

## 特性
- 完全本地运行大模型推理，不需要其他llm api
- 支持多种模型，包括qewn、llama2、gemma、mixtral等(具体模型请访问[Ollama Library](https://ollama.com/library))
- 一键启动，无需其他关联服务

## 架构图
![](doc/doc.drawio.svg)

## 使用
TODO

## 编译
```bash
make build
```