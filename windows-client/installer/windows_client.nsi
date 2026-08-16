; ============================================================
; 系统加固 Windows 客户端安装包脚本
; 编译：makensis windows_client.nsi
; 产物：dist/SystemHardening_WindowsClient_Setup_1.0.0.exe
; ============================================================

!define APP_NAME "系统加固 Windows 客户端"
!define APP_VERSION "2.1.6"
!define APP_EXE "windows_hardening_client.exe"
!define SERVICE_NAME "SystemHardeningWinClient"
!define UNINST_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\SystemHardeningWinClient"

Name "${APP_NAME}"
OutFile "..\..\dist_win\SystemHardening_WindowsClient_Setup_${APP_VERSION}.exe"
InstallDir "$PROGRAMFILES64\SystemHardening\WindowsClient"
RequestExecutionLevel admin
Unicode true

; 界面语言
!include "MUI2.nsh"
!include "LogicLib.nsh"
!include "x64.nsh"
!include "nsDialogs.nsh"

!define MUI_ABORTWARNING
!define MUI_ICON "${NSISDIR}\Contrib\Graphics\Icons\modern-install.ico"
!define MUI_UNICON "${NSISDIR}\Contrib\Graphics\Icons\modern-uninstall.ico"

; 服务器地址变量
Var ServerUrl

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
Page custom ServerConfigPage ServerConfigPageLeave
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "SimpChinese"

; ============================================================
; 服务器配置页面
; ============================================================
Function ServerConfigPage
    nsDialogs::Create 1018
    Pop $0

    ${NSD_CreateLabel} 0 0 100% 12u "请输入管理服务器地址："
    Pop $0

    ${NSD_CreateText} 0 15u 100% 12u "http://10.66.254.155:8080"
    Pop $ServerUrl

    nsDialogs::Show
FunctionEnd

Function ServerConfigPageLeave
    ${NSD_GetText} $ServerUrl $ServerUrl
FunctionEnd

; ============================================================
; 安装段
; ============================================================
Section "Install"
    ; 停止并删除旧服务（如果存在）
    ; 等待服务完全停止后再删除，避免文件被锁定导致复制失败
    nsExec::ExecToLog 'net stop "${SERVICE_NAME}"'
    Sleep 3000
    nsExec::ExecToLog 'sc delete "${SERVICE_NAME}"'
    Sleep 2000

    ; 创建安装目录和数据目录
    SetOutPath "$PROGRAMFILES64\SystemHardening\WindowsClient"
    SetOverwrite on
    File "..\target\x86_64-pc-windows-gnu\release\${APP_EXE}"

    ; 配置文件：仅全新安装时写入模板（更新时保留已有配置，避免 server_url 丢失）
    CreateDirectory "C:\ProgramData\SystemHardening\WindowsClient"
    IfFileExists "C:\ProgramData\SystemHardening\WindowsClient\config.yaml" SkipConfigWrite 0
        ; 全新安装：写入配置模板并替换服务器地址
        SetOutPath "C:\ProgramData\SystemHardening\WindowsClient"
        File /oname=config.yaml "..\config.example.yaml"
        SetOutPath "$PROGRAMFILES64\SystemHardening\WindowsClient"
        nsExec::ExecToLog 'powershell -Command "(Get-Content \"C:\ProgramData\SystemHardening\WindowsClient\config.yaml\" -Raw) -replace \"http://10.66.254.155:8080\",\"$ServerUrl\" | Set-Content \"C:\ProgramData\SystemHardening\WindowsClient\config.yaml\" -Encoding UTF8"'
        DetailPrint "Config file generated: C:\ProgramData\SystemHardening\WindowsClient\config.yaml"
    SkipConfigWrite:
    DetailPrint "Existing config preserved: C:\ProgramData\SystemHardening\WindowsClient\config.yaml"

    ; 创建卸载程序
    WriteUninstaller "$PROGRAMFILES64\SystemHardening\WindowsClient\uninstall.exe"

    ; Register service
    nsExec::ExecToLog 'sc create "${SERVICE_NAME}" binPath= "\"$PROGRAMFILES64\SystemHardening\WindowsClient\${APP_EXE}\"" start= auto DisplayName= "System Hardening Windows Client"'
    nsExec::ExecToLog 'sc description "${SERVICE_NAME}" "Collect Windows system hardening check data and report to management platform (read-only)"'
    nsExec::ExecToLog 'sc failure "${SERVICE_NAME}" reset= 86400 actions= restart/5000/restart/30000/restart/60000'
    nsExec::ExecToLog 'net start "${SERVICE_NAME}"'
    
    ; 写入卸载注册表项
    WriteRegStr HKLM "${UNINST_KEY}" "DisplayName" "${APP_NAME}"
    WriteRegStr HKLM "${UNINST_KEY}" "DisplayVersion" "${APP_VERSION}"
    WriteRegStr HKLM "${UNINST_KEY}" "Publisher" "SystemHardening"
    WriteRegStr HKLM "${UNINST_KEY}" "UninstallString" "$PROGRAMFILES64\SystemHardening\WindowsClient\uninstall.exe"
    WriteRegDWORD HKLM "${UNINST_KEY}" "NoModify" 1
    WriteRegDWORD HKLM "${UNINST_KEY}" "NoRepair" 1
    
    ; Installation complete message
    MessageBox MB_ICONINFORMATION|MB_OK "Installation complete!$\r$\n$\r$\nService registered and started.$\r$\nServer address: $ServerUrl"
SectionEnd

; ============================================================
; 卸载段
; ============================================================
Section "Uninstall"
    ; 停止并删除服务
    nsExec::ExecToLog 'net stop "${SERVICE_NAME}"'
    nsExec::ExecToLog 'sc delete "${SERVICE_NAME}"'

    ; 删除文件
    Delete "$PROGRAMFILES64\SystemHardening\WindowsClient\${APP_EXE}"
    Delete "$PROGRAMFILES64\SystemHardening\WindowsClient\uninstall.exe"
    RMDir "$PROGRAMFILES64\SystemHardening\WindowsClient"

    ; 删除卸载注册表项
    DeleteRegKey HKLM "${UNINST_KEY}"

    ; 删除数据目录（Tokens 和配置文件）
    Delete "C:\ProgramData\SystemHardening\WindowsClient\tokens.json"
    Delete "C:\ProgramData\SystemHardening\WindowsClient\config.yaml"
    RMDir "C:\ProgramData\SystemHardening\WindowsClient"
    RMDir "C:\ProgramData\SystemHardening"

    ; 卸载完成提示
    MessageBox MB_ICONINFORMATION|MB_OK "卸载完成。$\r$\n客户端已从系统中移除。"
SectionEnd
