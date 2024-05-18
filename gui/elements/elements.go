package elements

type GuiElements struct {
   AuthOffline   *AuthOffline
   Auth          *Auth
   AccountList   *AccountList
   Settings      *Settings
   ProfileList   *ProfileList
   ProfileEdit   *ProfileEdit
   HomeScreen    *Home
   CrashInformer *CrashInformer
}

func New() *GuiElements {
   return &GuiElements{
      AuthOffline: NewAuthOffline(),
      Auth: NewAuth(),
      AccountList: NewAccountList(),
      Settings: NewSettings(),
      ProfileList: NewProfileList(),
      ProfileEdit: NewProfileEdit(),
      HomeScreen: NewHome(),
      CrashInformer: NewCrashInformer(),
   }
}
