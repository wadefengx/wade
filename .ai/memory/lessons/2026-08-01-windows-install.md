# 经验教训

## 2026-08-01 Windows 安装体验

### L9: Windows 缺一键安装路径是最大断点
macOS/Linux 有 install.sh,Windows 只有 scoop(需先装 scoop,且安装脚本是 PowerShell 的,cmd 用户直接 '不是内部或外部命令')。
**教训**: 跨平台工具必须有对称的安装体验 —— 补了 install.ps1(irm ... | iex),cmd/PowerShell 通用,自动加用户 PATH。

### L10: scoop post_install 里不能跑交互式命令
`"wade.exe setup"` 无 --auto 时等输入,scoop 安装会卡死。
**教训**: manifest post_install 必须非交互:`wade.exe setup --auto`。

### L11: 用户画像模拟找断点
"自己 Windows 上用着费劲" → 逐角色模拟(纯 cmd 零基础/有 scoop/不想装 scoop) → 定位每个 '不是内部或外部命令' 的真实根因。
**教训**: 文档给的命令假设用户已有前置工具;要么补前置安装步骤,要么给一条不依赖前置的路径。
