// import * as S5Conf from '../conf';

export enum fontWeight {
  Normal = '400',
  Bold = '700',
  Light = '100',
  Black = '900',
}

export enum fontStyle {
  Regular = 'normal',
  Italic = 'italic',
}

// like color, this is just a way to have types when the underlying
// true object is just a string
export default interface font {
  Family: string;
  Weight: fontWeight;
  Style: fontStyle;
  Size: number;
}

export interface fontFamily {
  Font(s: fontStyle, w: fontWeight, z: number): font;
}

export const browserSerif: fontFamily = {
  Font: (s: fontStyle, w: fontWeight): font => {
    return {
      Family: 'Serif',
      Style: s,
      Weight: w,
      Size: 12,
    };
  },
};

export const browserSansSerif: fontFamily = {
  Font: (s: fontStyle, w: fontWeight): font => {
    return {
      Family: 'SansSerif',
      Style: s,
      Weight: w,
      Size: 12,
    };
  },
};

export const fontHeight = (f: font): number => {
  return f.Size;
};
