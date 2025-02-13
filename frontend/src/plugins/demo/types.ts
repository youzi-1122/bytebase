type StepType = "click" | "change";

export type DialogPosition = "bottom" | "top" | "left" | "right" | "topright";

export type I18NText = {
  [key: string]: string;
};

export interface StepData {
  type: StepType;
  title: string | I18NText;
  description: string | I18NText;
  selectors: string[][];
  // url is using for validate url in change step
  url?: string;
  // value is the regex-like string using for check the target content value
  value?: string;
  // position is the position of the guide dialog (default is bottom)
  position?: DialogPosition;
  // cover is the flag that cover should be shown
  cover?: boolean;
  // hideNextButton is the flag that next button should be hidden
  hideNextButton?: boolean;
}

export interface GuideData {
  name: string;
  steps: StepData[];
}

export type HintType = "hint" | "shield";

// Hint is a special guide that has no Next button and is always shown.
export interface HintData {
  selector: string;
  type: HintType;
  // pathname is the wanted pathname of the url, cound be a regex-like string
  pathname: string;
  // url is using for validate url in change step
  url: string;
  highlight?: boolean;
  // cover is the flag that cover should be shown
  cover?: boolean;
  // dialog is a data of dialog info. If it's undefined, then the dialog will not be shown.
  dialog?: {
    title: string | I18NText;
    description: string | I18NText;
    // position is the position of the guide dialog (default is bottom)
    position?: DialogPosition;
    alwaysShow?: boolean;
    showOnce?: boolean;
  };
  // addtionClass is the class name that should be added to the hint element
  additionStyle?: CSSStyleDeclaration;
}
