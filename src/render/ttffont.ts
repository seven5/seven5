// I want both of these classes in this file because they are brothers
/* eslint-disable max-classes-per-file */

import * as S5Conf from '../conf';
import font, { fontFamily, fontStyle, fontWeight } from './font';

export default class ttfFont implements font {
  Family: string;

  Weight: fontWeight = fontWeight.Normal;

  Style: fontStyle = fontStyle.Regular;

  Size = 12;

  constructor(family: string) {
    this.Family = family;
  }
}

export class ttfFontFamily implements fontFamily {
  family: string;

  constructor(f: string) {
    this.family = f;
  }

  // urlPrefix is prefix in URL space to load from, with trailing /
  // for no prefix, pass ""
  //
  // note that this expects the filenames of the font to be 8 names
  // exactly like this
  //
  // foo-Regular.ttf , foo-Italic.ttf
  // foo-Light.ttf, foo-LightItalic.ttf
  // foo-Bold.ttf foo-BoldItalic.ttf
  // foo-Black.ttf foo-BlackItalic.ttf

  static Load = (name: string, urlPrefix: string): Promise<ttfFontFamily> => {
    const suffixes = [
      'Regular',
      'Italic',
      'Light',
      'LightItalic',
      'Bold',
      'BoldItalic',
      'Black',
      'BlackItalic',
    ];
    const styles = [
      'normal',
      'italic',
      'normal',
      'italic',
      'normal',
      'italic',
      'normal',
      'italic',
    ];
    const weights = ['450', '450', '100', '100', '700', '700', '900', '900'];
    const promises = [] as Promise<FontFace>[];
    // because eslint doesn't like for loops!
    const indices = [0, 1, 2, 3, 4, 5, 6, 7];
    indices.forEach((index: number) => {
      const suffix = suffixes[index];
      const style = styles[index];
      const weight = weights[index];
      const f = new FontFace(
        name,
        `url('${urlPrefix}Merriweather-${suffix}.ttf')`,
        {
          weight,
          style,
        }
      );
      promises.push(f.load());
    });
    return new Promise<ttfFontFamily>((resolve, reject) => {
      Promise.all(promises)
        .then((ff: FontFace[]) => {
          ff.forEach((f: FontFace) => document.fonts.add(f));
          resolve(new ttfFontFamily(name));
        })
        .catch((reason: unknown) => {
          S5Conf.Error(`failed to load font family "${name}":${reason}`);
          reject(reason);
        });
    });
  };

  Font(s: fontStyle, w: fontWeight, size: number): font {
    const result = new ttfFont(this.family);
    result.Weight = w;
    result.Style = s;
    result.Size = size;
    return result;
  }
}
