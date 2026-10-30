import os
import uuid
from pathlib import Path
from flask import Flask, render_template, request, redirect, url_for, flash, send_from_directory, abort
from PIL import Image, UnidentifiedImageError

app = Flask(__name__)
app.secret_key = os.environ.get('SECRET_KEY', 'local-demo-only-change-me')
UPLOAD_DIR = Path(os.environ.get('UPLOAD_DIR', '/app/uploads'))
UPLOAD_DIR.mkdir(parents=True, exist_ok=True)
MAX_SIZE = 10 * 1024 * 1024
app.config['MAX_CONTENT_LENGTH'] = MAX_SIZE

@app.get('/healthz')
def health():
    return {'status': 'ok'}

@app.get('/')
def index():
    images = sorted((p.name for p in UPLOAD_DIR.iterdir() if p.is_file() and p.suffix.lower() in ('.jpg', '.png', '.webp')), reverse=True)
    return render_template('index.html', images=images)

@app.post('/upload')
def upload():
    file = request.files.get('picture')
    if not file or not file.filename:
        flash('Choose an image first.')
        return redirect(url_for('index'))
    try:
        img = Image.open(file.stream)
        img.verify()
        file.stream.seek(0)
        img = Image.open(file.stream)
        if img.format not in ('JPEG', 'PNG', 'WEBP'):
            raise ValueError('Unsupported image format')
        fmt = img.format
        ext = {'JPEG': '.jpg', 'PNG': '.png', 'WEBP': '.webp'}[fmt]
        filename = f'{uuid.uuid4().hex}{ext}'
        # Decode and re-encode to avoid storing arbitrary uploaded bytes.
        img.load()
        if img.width * img.height > 25_000_000:
            raise ValueError('Image dimensions are too large')
        if fmt == 'JPEG' and img.mode not in ('RGB', 'L'):
            img = img.convert('RGB')
        img.save(UPLOAD_DIR / filename, format=fmt)
        flash('Picture uploaded to persistent storage.')
    except (UnidentifiedImageError, OSError, ValueError, Image.DecompressionBombError):
        flash('Invalid picture. Upload a JPG, PNG, or WebP under 10 MB.')
    return redirect(url_for('index'))

@app.get('/images/<path:filename>')
def images(filename):
    if Path(filename).name != filename or Path(filename).suffix.lower() not in ('.jpg', '.png', '.webp'):
        abort(404)
    return send_from_directory(UPLOAD_DIR, filename)

@app.errorhandler(413)
def too_large(error):
    return 'Image exceeds 10 MB limit', 413

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)
