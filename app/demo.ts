import * as S5Interactor from '../src/interactor';
import * as S5Conf from '../src/conf';
import * as S5Input from '../src/input';

import {
  Font,
  TTFFontFamily,
  FontFamily,
  FontWeight,
  FontStyle,
  DefaultTheme,
} from '../src/render';

const label = (text: string, f: Font): S5Interactor.Label => {
  const l = new S5Interactor.Label(text, 'label');
  l.Font = f;
  l.TransparentBackground = true;
  return l;
};

const init = () => {
  const r = new S5Interactor.Root();
  const ff = DefaultTheme.FontFamily;
  const col = new S5Interactor.Column();
  const pm = S5Input.CreateInputPolicyManager();

  // initialize input system with default policies/agents... pull events
  // from the root HTML element
  pm.InitializePoliciesAndDispatchAgents(r);

  const l0 = label(
    'This is a column interactor:',
    ff.Font(FontStyle.Regular, FontWeight.Normal, 32)
  );
  col.AddChild(l0);
  const l1 = label(
    'The quick brown fox',
    ff.Font(FontStyle.Regular, FontWeight.Bold, 32)
  );
  col.AddChild(l1);
  const l2 = label(
    'jumps over the',
    ff.Font(FontStyle.Regular, FontWeight.Light, 32)
  );
  col.AddChild(l2);
  const l3 = label(
    'lazy dog.',
    ff.Font(FontStyle.Regular, FontWeight.Black, 32)
  );
  col.AddChild(l3);

  r.AddChildAt(col, 50, 100);
  const b0 = new S5Interactor.Button('example button', 'button');
  b0.TransparentBackground = false;
  b0.Font = ff.Font(FontStyle.Regular, FontWeight.Black, 32);

  r.AddChildAt(b0, 400, 150);

  r.StartMainloop(true);
};

const fontPreLoad = (): Promise<TTFFontFamily> => {
  return new Promise<TTFFontFamily>((resolve, reject) => {
    const p = TTFFontFamily.Load('merriweather', '');
    p.then((fam: TTFFontFamily) => {
      resolve(fam);
    }).catch((reason) => {
      S5Conf.Error(
        'Unable to load merriweather, falling back to browser default fonts'
      );
      reject(reason);
    });
  });
};

fontPreLoad()
  .then((fam: TTFFontFamily) => {
    DefaultTheme.FontFamily = fam as FontFamily;
    init();
  })
  .catch(() => {
    init();
  }); // calls init whether fonts are ready or not with DefaultTheme.FontFamily
