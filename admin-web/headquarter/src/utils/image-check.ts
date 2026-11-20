export interface PngRule {
  width: number;
  height: number;
  maxSize: number;
  /** 头像框：中心像素必须完全透明，且至少有一个非全透明像素 */
  transparentCenter?: boolean;
}

export type PngCheckError =
  | 'type'
  | 'size'
  | 'decode'
  | 'dimension'
  | 'centerOpaque'
  | 'allTransparent';

const PNG_SIGNATURE = [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a];

async function hasPngSignature(file: File) {
  const head = new Uint8Array(await file.slice(0, 8).arrayBuffer());
  return PNG_SIGNATURE.every((byte, i) => head[i] === byte);
}

function loadImage(file: File) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const url = URL.createObjectURL(file);
    const img = new Image();
    img.onload = () => {
      URL.revokeObjectURL(url);
      resolve(img);
    };
    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error('decode'));
    };
    img.src = url;
  });
}

function checkTransparency(img: HTMLImageElement): PngCheckError | null {
  const { naturalWidth: width, naturalHeight: height } = img;
  const canvas = document.createElement('canvas');
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext('2d', { willReadFrequently: true });
  if (!ctx) return null;
  ctx.drawImage(img, 0, 0);
  const { data } = ctx.getImageData(0, 0, width, height);
  const centerAlpha =
    data[(Math.floor(height / 2) * width + Math.floor(width / 2)) * 4 + 3];
  if (centerAlpha !== 0) return 'centerOpaque';
  for (let i = 3; i < data.length; i += 4) {
    if (data[i] !== 0) return null;
  }
  return 'allTransparent';
}

/** 上传前的本地预检，最终以服务端校验为准 */
export async function checkPngFile(
  file: File,
  rule: PngRule
): Promise<PngCheckError | null> {
  if (!/\.png$/i.test(file.name)) return 'type';
  if (file.size > rule.maxSize) return 'size';
  if (!(await hasPngSignature(file))) return 'type';
  let img: HTMLImageElement;
  try {
    img = await loadImage(file);
  } catch {
    return 'decode';
  }
  if (img.naturalWidth !== rule.width || img.naturalHeight !== rule.height) {
    return 'dimension';
  }
  return rule.transparentCenter ? checkTransparency(img) : null;
}
