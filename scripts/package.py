#!/usr/bin/env python3
"""Prepare a preview package without redistributing external game/map assets."""
from hashlib import sha256
from pathlib import Path
import json,struct,zipfile
ROOT=Path(__file__).resolve().parents[1]
RELEASE=ROOT/'dist'/'ArayMinimap-v0.1.0-preview.1-win64.zip'
SOURCES=ROOT/'dist'/'ArayMinimap-v0.1.0-preview.1-source.zip'
files=['README.md','LICENSE','install.ps1','install.cmd','uninstall.ps1','uninstall.cmd','scripts/import_tiles.py','dist/ArayMinimap.dll','dist/ArayMinimap.exe','dist/ArayMinimap-Check.exe']
for name in ['dist/ArayMinimap.dll','dist/ArayMinimap.exe','dist/ArayMinimap-Check.exe']:
 data=(ROOT/name).read_bytes()
 assert data[:2]==b'MZ',f'Not a PE: {name}'
 pos=struct.unpack_from('<I',data,0x3c)[0]
 assert data[pos:pos+4]==b'PE\0\0',f'PE signature invalid: {name}'
 assert struct.unpack_from('<H',data,pos+4)[0]==0x8664,f'Not AMD64: {name}'
manifest={'version':'0.1.0-preview.1','game_compatibility':'not validated','items':{}}
with zipfile.ZipFile(RELEASE,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=7) as z:
 for name in files:
  file=ROOT/name
  if not file.is_file():raise FileNotFoundError(str(file))
  body=file.read_bytes();manifest['items'][name]={'sha256':sha256(body).hexdigest(),'bytes':len(body)}
  z.writestr('ArayMinimap/'+name,body)
 z.writestr('ArayMinimap/SHA256SUMS.json',json.dumps(manifest,ensure_ascii=False,indent=2))
assert zipfile.ZipFile(RELEASE).testzip() is None
with zipfile.ZipFile(SOURCES,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=6) as z:
 for file in ROOT.rglob('*'):
  if not file.is_file() or any(p in ('dist','build','__pycache__') for p in file.relative_to(ROOT).parts):continue
  z.write(file,arcname='ArayMinimap/'+file.relative_to(ROOT).as_posix())
assert zipfile.ZipFile(SOURCES).testzip() is None
print('Preview:',RELEASE,RELEASE.stat().st_size,'bytes')
print('Sources:',SOURCES,SOURCES.stat().st_size,'bytes')
print('Testzip: OK; PE AMD64: OK; third-party map assets: not included')
