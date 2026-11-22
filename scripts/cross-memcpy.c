/* Cross-link shim for Clang/LLD PE toolchains WITHOUT Microsoft CRT headers.
 * This file is NOT compiled in the MSVC Windows pipeline, where /MT provides
 * a normal memcpy implementation. Never distribute this as a DLL.
 */
typedef unsigned long long aray_size_t;
void *memcpy(void *dest,const void *src,aray_size_t size){
 volatile unsigned char *d=(volatile unsigned char*)dest;
 const volatile unsigned char *s=(const volatile unsigned char*)src;
 for(aray_size_t i=0;i<size;++i)d[i]=s[i];
 return dest;
}
