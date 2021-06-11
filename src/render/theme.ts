import color from "./color";
import  {fontFamily} from './font';

interface theme {
  FontFamily: fontFamily;
  Foreground: color;
  Background: color;
  Emphasis: color;
  Error: color;
}

export default theme;
