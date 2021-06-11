import * as S5Err from '../error';

const loadBitmap = async (
  url: string,
  w: number,
  h: number
): Promise<ImageBitmap> => {
  return fetch(url)
    .then(
      (r: Response): Promise<Blob> => {
        if (r.status !== 200) {
          return Promise.reject(
            S5Err.NewError(
              S5Err.Messages.NetworkError,
              `Status: ${r.status}, StatusText: ${r.statusText}]`
            )
          );
        }
        return r.blob();
      }
    )
    .then(
      (blob: Blob): Promise<ImageBitmap> => {
        return createImageBitmap(blob, 0, 0, w, h);
      }
    )
    .catch(
      (reason: unknown): Promise<ImageBitmap> => {
        return Promise.reject(reason);
      }
    );
};

export default loadBitmap;
