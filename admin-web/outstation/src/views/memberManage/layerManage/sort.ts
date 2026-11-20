import type { LayerRecord } from './types';

/** 自动层级按晋升门槛从低到高排列，默认层级始终在最前。 */
export default function sortAutoLayers(list: LayerRecord[]) {
  return [...list].sort((a, b) => {
    if (a.isDefault !== b.isDefault) return a.isDefault ? -1 : 1;
    return (a.deposit ?? 0) - (b.deposit ?? 0) || (a.wash ?? 0) - (b.wash ?? 0);
  });
}
