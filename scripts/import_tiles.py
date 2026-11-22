#!/usr/bin/env python3
"""Convert locally obtained JW map tiles to formats supported by stdlib Go.

The original resources are NOT distributed by Aray Minimap. Only use files
that you are legally permitted to obtain and keep for your own use.
Requires: python -m pip install Pillow
"""
from argparse import ArgumentParser
from pathlib import Path

def main():
    ap=ArgumentParser(description='Importe tiles WEBP locais para PNG/JPEG')
    ap.add_argument('--source',required=True,help='Pasta ctiles do mapa obtido pelo usuario')
    ap.add_argument('--destination',required=True,help='Pasta data/ctiles do Aray Minimap')
    ap.add_argument('--max-tiles',type=int,default=0,help='Limite opcional para testes')
    args=ap.parse_args()
    try: from PIL import Image
    except ImportError as e: raise SystemExit('Instale Pillow: python -m pip install Pillow') from e
    root=Path(args.source).resolve();target=Path(args.destination).resolve()
    if not root.is_dir(): raise SystemExit('Origem nao existe: '+str(root))
    count=0
    for file in sorted(root.rglob('*.webp')):
        rel=file.relative_to(root)
        out=target/rel.with_suffix('.png')
        if out.is_file(): continue
        out.parent.mkdir(parents=True,exist_ok=True)
        with Image.open(file) as im: im.convert('RGB').save(out,format='PNG',optimize=True)
        count+=1
        if args.max_tiles and count>=args.max_tiles: break
    print(f'Tiles convertidos: {count} | destino: {target}')
if __name__=='__main__':main()
