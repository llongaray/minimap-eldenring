/* Aray Minimap -- independent Elden Mod Loader bridge, Windows x64.
 * SPDX-License-Identifier: GPL-3.0-only
 * No third-party source files or hook libraries are linked into this DLL.
 * DllMain only schedules work; CreateProcess happens on a separate thread.
 */
#define ARAY_EXPORT __declspec(dllexport)
#define ARAY_IMPORT __declspec(dllimport)
#define WINAPI __stdcall
#define NULL ((void*)0)
typedef void* HANDLE;
typedef void* HMODULE;
typedef unsigned long DWORD;
typedef unsigned short WCHAR;
typedef int BOOL;
typedef unsigned long long SIZE_T;
typedef DWORD (WINAPI *LPTHREAD_START_ROUTINE)(void*);
typedef struct { DWORD cb; WCHAR* reserved; WCHAR* desktop; WCHAR* title;
 DWORD dwX,dwY,dwXSize,dwYSize,dwXCountChars,dwYCountChars,dwFillAttribute,dwFlags;
 unsigned short wShowWindow,cbReserved2; unsigned char* lpReserved2;
 HANDLE hStdInput,hStdOutput,hStdError;
} STARTUPINFOW;
typedef struct { HANDLE hProcess,hThread; DWORD dwProcessId,dwThreadId; } PROCESS_INFORMATION;
ARAY_IMPORT DWORD WINAPI GetModuleFileNameW(HMODULE,WCHAR*,DWORD);
ARAY_IMPORT BOOL WINAPI DisableThreadLibraryCalls(HMODULE);
ARAY_IMPORT HANDLE WINAPI CreateThread(void*,SIZE_T,LPTHREAD_START_ROUTINE,void*,DWORD,DWORD*);
ARAY_IMPORT BOOL WINAPI CloseHandle(HANDLE);
ARAY_IMPORT BOOL WINAPI CreateProcessW(const WCHAR*,WCHAR*,void*,void*,BOOL,DWORD,void*,const WCHAR*,STARTUPINFOW*,PROCESS_INFORMATION*);

/* Avoid linking the CRT. The compiler may emit memcpy for POD zero-inits. */
void *memcpy(void *dest, const void *src, SIZE_T size){
 volatile unsigned char *d=(volatile unsigned char*)dest;
 const volatile unsigned char *s=(const volatile unsigned char*)src;
 for(SIZE_T i=0;i<size;i++)d[i]=s[i];
 return dest;
}
static WCHAR module_path[32768], exe_path[32768], command[32768];
static DWORD WINAPI launch_worker(void* module){
 DWORD len=GetModuleFileNameW((HMODULE)module,module_path,32768);
 if(!len||len>=32768-100)return 1;
 // Module folder: Game/mods. The bridge is the only file directly in mods/.
 DWORD sep=len;
 while(sep>0&&module_path[sep-1]!='\\'&&module_path[sep-1]!='/')sep--;
 if(sep==0||sep>32768-150)return 2;
 // String building without any C runtime dependency.
 static const WCHAR suffix[]={ 'A','r','a','y','M','i','n','i','m','a','p','\\','A','r','a','y','M','i','n','i','m','a','p','.','e','x','e',0};
 DWORD i=0;
 for(i=0;i<sep;i++)exe_path[i]=module_path[i];
 for(DWORD j=0;suffix[j];j++)exe_path[i++]=suffix[j];
 exe_path[i]=0;
 // Quoted executable path; the helper derives its data directory from the
 // companion location and does not depend on the working directory.
 DWORD k=0;command[k++]='"';
 for(DWORD j=0;j<i;j++)command[k++]=exe_path[j];
 command[k++]='"';command[k]=0;
 STARTUPINFOW si={0};PROCESS_INFORMATION pi={0};si.cb=(DWORD)sizeof(si);
 if(!CreateProcessW(exe_path,command,NULL,NULL,0,0x08000000,NULL,NULL,&si,&pi))return 3;
 CloseHandle(pi.hThread);CloseHandle(pi.hProcess);
 return 0;
}
BOOL WINAPI DllMain(HMODULE module,DWORD reason,void* reserved){
 (void)reserved;
 if(reason==1){
  DisableThreadLibraryCalls(module);
  HANDLE thread=CreateThread(NULL,0,launch_worker,module,0,NULL);
  if(thread)CloseHandle(thread);
 }
 return 1;
}
ARAY_EXPORT DWORD WINAPI ArayMinimapInterfaceVersion(void){return 1;}
