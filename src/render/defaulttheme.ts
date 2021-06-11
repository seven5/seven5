import color, { rgbColor } from './color';
import { browserSerif, fontFamily } from './font';
import theme from './theme';

const defaultTheme = {
  FontFamily: browserSerif as fontFamily,
  Foreground: new rgbColor(0, 0, 0) as color,
  Background: new rgbColor(0xf9, 0xf9, 0xf9) as color,
  Emphasis: new rgbColor(0x4b, 0x78, 0xc5) as color,
  Error: new rgbColor(0xc5, 0x5b, 0x4b) as color,
};

export default defaultTheme as theme;
