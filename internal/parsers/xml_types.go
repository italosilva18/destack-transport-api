package parsers

import (
	"encoding/xml"
	"time"
)

// ============ Estruturas para CT-e ============

// CTeProc representa o processo do CT-e completo
type CTeProc struct {
	XMLName xml.Name `xml:"cteProc"`
	Versao  string   `xml:"versao,attr"`
	CTe     CTe      `xml:"CTe"`
	ProtCTe ProtCTe  `xml:"protCTe"`
}

// CTe representa o CT-e
type CTe struct {
	XMLName    xml.Name   `xml:"CTe"`
	InfCte     InfCte     `xml:"infCte"`
	InfCTeSupl InfCTeSupl `xml:"infCTeSupl"`
	Signature  Signature  `xml:"Signature"`
}

// InfCte contém as informações do CT-e
type InfCte struct {
	XMLName    xml.Name   `xml:"infCte"`
	Versao     string     `xml:"versao,attr"`
	Id         string     `xml:"Id,attr"`
	Ide        IdeCTe     `xml:"ide"`
	Compl      ComplCTe   `xml:"compl"`
	Emit       Emit       `xml:"emit"`
	Rem        Rem        `xml:"rem"`
	Dest       Dest       `xml:"dest"`
	VPrest     VPrest     `xml:"vPrest"`
	Imp        ImpCTe     `xml:"imp"`
	InfCTeNorm InfCTeNorm `xml:"infCTeNorm"`
	InfRespTec InfRespTec `xml:"infRespTec"`
}

// IdeCTe identificação do CT-e
type IdeCTe struct {
	CUF       string `xml:"cUF"`
	CCT       string `xml:"cCT"`
	CFOP      string `xml:"CFOP"`
	NatOp     string `xml:"natOp"`
	Mod       string `xml:"mod"`
	Serie     string `xml:"serie"`
	NCT       string `xml:"nCT"`
	DhEmi     string `xml:"dhEmi"`
	TpImp     string `xml:"tpImp"`
	TpEmis    string `xml:"tpEmis"`
	CDV       string `xml:"cDV"`
	TpAmb     string `xml:"tpAmb"`
	TpCTe     string `xml:"tpCTe"`
	ProcEmi   string `xml:"procEmi"`
	VerProc   string `xml:"verProc"`
	CMunEnv   string `xml:"cMunEnv"`
	XMunEnv   string `xml:"xMunEnv"`
	UFEnv     string `xml:"UFEnv"`
	Modal     string `xml:"modal"`
	TpServ    string `xml:"tpServ"`
	CMunIni   string `xml:"cMunIni"`
	XMunIni   string `xml:"xMunIni"`
	UFIni     string `xml:"UFIni"`
	CMunFim   string `xml:"cMunFim"`
	XMunFim   string `xml:"xMunFim"`
	UFFim     string `xml:"UFFim"`
	Retira    string `xml:"retira"`
	IndIEToma string `xml:"indIEToma"`
	Toma3     Toma3  `xml:"toma3"`
}

// ComplCTe complemento do CT-e
type ComplCTe struct {
	XEmi    string    `xml:"xEmi"`
	XObs    string    `xml:"xObs"`
	ObsCont []ObsCont `xml:"ObsCont"`
}

// ObsCont observações do contribuinte
type ObsCont struct {
	XCampo string `xml:"xCampo,attr"`
	XTexto string `xml:"xTexto"`
}

// Emit emitente
type Emit struct {
	CNPJ      string   `xml:"CNPJ"`
	IE        string   `xml:"IE"`
	XNome     string   `xml:"xNome"`
	XFant     string   `xml:"xFant"`
	EnderEmit Endereco `xml:"enderEmit"`
	CRT       string   `xml:"CRT"`
}

// Rem remetente
type Rem struct {
	CNPJ      string   `xml:"CNPJ"`
	CPF       string   `xml:"CPF"`
	IE        string   `xml:"IE"`
	XNome     string   `xml:"xNome"`
	EnderReme Endereco `xml:"enderReme"`
}

// Dest destinatário
type Dest struct {
	CNPJ      string   `xml:"CNPJ"`
	CPF       string   `xml:"CPF"`
	IE        string   `xml:"IE"`
	XNome     string   `xml:"xNome"`
	Fone      string   `xml:"fone"`
	EnderDest Endereco `xml:"enderDest"`
}

// Endereco endereço genérico
type Endereco struct {
	XLgr    string `xml:"xLgr"`
	Nro     string `xml:"nro"`
	XCpl    string `xml:"xCpl"`
	XBairro string `xml:"xBairro"`
	CMun    string `xml:"cMun"`
	XMun    string `xml:"xMun"`
	CEP     string `xml:"CEP"`
	UF      string `xml:"UF"`
	CPais   string `xml:"cPais"`
	XPais   string `xml:"xPais"`
	Email   string `xml:"email"`
}

// VPrest valores da prestação
type VPrest struct {
	VTPrest string `xml:"vTPrest"`
	VRec    string `xml:"vRec"`
	Comp    []Comp `xml:"Comp"`
}

// Comp componente do valor
type Comp struct {
	XNome string `xml:"xNome"`
	VComp string `xml:"vComp"`
}

// InfCTeNorm informações normais do CT-e
type InfCTeNorm struct {
	InfCarga InfCarga `xml:"infCarga"`
	InfDoc   InfDoc   `xml:"infDoc"`
	InfModal InfModal `xml:"infModal"`
}

// InfCarga informações da carga
type InfCarga struct {
	VCarga      string `xml:"vCarga"`
	ProPred     string `xml:"proPred"`
	InfQ        []InfQ `xml:"infQ"`
	VCargaAverb string `xml:"vCargaAverb"`
}

// InfQ quantidade
type InfQ struct {
	CUnid  string `xml:"cUnid"`
	TpMed  string `xml:"tpMed"`
	QCarga string `xml:"qCarga"`
}

// InfDoc documentos
type InfDoc struct {
	InfNFe []InfNFe `xml:"infNFe"`
}

// InfNFe informações da NFe
type InfNFe struct {
	Chave string `xml:"chave"`
}

// ProtCTe protocolo de autorização
type ProtCTe struct {
	InfProt InfProt `xml:"infProt"`
}

// InfProt informações do protocolo
type InfProt struct {
	TpAmb    string `xml:"tpAmb"`
	VerAplic string `xml:"verAplic"`
	ChCTe    string `xml:"chCTe"`
	DhRecbto string `xml:"dhRecbto"`
	NProt    string `xml:"nProt"`
	DigVal   string `xml:"digVal"`
	CStat    string `xml:"cStat"`
	XMotivo  string `xml:"xMotivo"`
}

// ============ Estruturas para MDF-e ============

// MDFeProc processo do MDF-e
type MDFeProc struct {
	XMLName  xml.Name `xml:"mdfeProc"`
	Versao   string   `xml:"versao,attr"`
	MDFe     MDFe     `xml:"MDFe"`
	ProtMDFe ProtMDFe `xml:"protMDFe"`
}

// MDFe manifesto eletrônico
type MDFe struct {
	XMLName     xml.Name    `xml:"MDFe"`
	InfMDFe     InfMDFe     `xml:"infMDFe"`
	InfMDFeSupl InfMDFeSupl `xml:"infMDFeSupl"`
	Signature   Signature   `xml:"Signature"`
}

// InfMDFe informações do MDF-e
type InfMDFe struct {
	Versao     string       `xml:"versao,attr"`
	Id         string       `xml:"Id,attr"`
	Ide        IdeMDFe      `xml:"ide"`
	Emit       Emit         `xml:"emit"`
	InfModal   InfModalMDFe `xml:"infModal"`
	InfDoc     InfDocMDFe   `xml:"infDoc"`
	Seg        []Seg        `xml:"seg"`
	ProdPred   ProdPred     `xml:"prodPred"`
	Tot        TotMDFe      `xml:"tot"`
	InfAdic    InfAdic      `xml:"infAdic"`
	InfRespTec InfRespTec   `xml:"infRespTec"`
}

// IdeMDFe identificação do MDF-e
type IdeMDFe struct {
	CUF           string        `xml:"cUF"`
	TpAmb         string        `xml:"tpAmb"`
	TpEmit        string        `xml:"tpEmit"`
	TpTransp      string        `xml:"tpTransp"`
	Mod           string        `xml:"mod"`
	Serie         string        `xml:"serie"`
	NMDF          string        `xml:"nMDF"`
	CMDF          string        `xml:"cMDF"`
	CDV           string        `xml:"cDV"`
	Modal         string        `xml:"modal"`
	DhEmi         string        `xml:"dhEmi"`
	TpEmis        string        `xml:"tpEmis"`
	ProcEmi       string        `xml:"procEmi"`
	VerProc       string        `xml:"verProc"`
	UFIni         string        `xml:"UFIni"`
	UFFim         string        `xml:"UFFim"`
	InfMunCarrega InfMunCarrega `xml:"infMunCarrega"`
}

// InfMunCarrega município de carregamento
type InfMunCarrega struct {
	CMunCarrega string `xml:"cMunCarrega"`
	XMunCarrega string `xml:"xMunCarrega"`
}

// InfModalMDFe modal do MDF-e
type InfModalMDFe struct {
	VersaoModal string `xml:"versaoModal,attr"`
	Rodo        Rodo   `xml:"rodo"`
}

// Rodo modal rodoviário
type Rodo struct {
	InfANTT    InfANTT    `xml:"infANTT"`
	VeicTracao VeicTracao `xml:"veicTracao"`
}

// InfANTT informações da ANTT
type InfANTT struct {
	RNTRC          string           `xml:"RNTRC"`
	InfContratante []InfContratante `xml:"infContratante"`
}

// InfContratante contratante
type InfContratante struct {
	CNPJ string `xml:"CNPJ"`
	CPF  string `xml:"CPF"`
}

// VeicTracao veículo de tração
type VeicTracao struct {
	Placa    string     `xml:"placa"`
	RENAVAM  string     `xml:"RENAVAM"`
	Tara     string     `xml:"tara"`
	CapKG    string     `xml:"capKG"`
	Prop     Prop       `xml:"prop"`
	Condutor []Condutor `xml:"condutor"`
	TpRod    string     `xml:"tpRod"`
	TpCar    string     `xml:"tpCar"`
	UF       string     `xml:"UF"`
}

// Prop proprietário
type Prop struct {
	CNPJ   string `xml:"CNPJ"`
	CPF    string `xml:"CPF"`
	RNTRC  string `xml:"RNTRC"`
	XNome  string `xml:"xNome"`
	IE     string `xml:"IE"`
	UF     string `xml:"UF"`
	TpProp string `xml:"tpProp"`
}

// Condutor condutor do veículo
type Condutor struct {
	XNome string `xml:"xNome"`
	CPF   string `xml:"CPF"`
}

// InfDocMDFe documentos do MDF-e
type InfDocMDFe struct {
	InfMunDescarga []InfMunDescarga `xml:"infMunDescarga"`
}

// InfMunDescarga município de descarga
type InfMunDescarga struct {
	CMunDescarga string   `xml:"cMunDescarga"`
	XMunDescarga string   `xml:"xMunDescarga"`
	InfCTe       []InfCTe `xml:"infCTe"`
	InfNFe       []InfNFe `xml:"infNFe"`
}

// InfCTe CT-e vinculado
type InfCTe struct {
	ChCTe string `xml:"chCTe"`
}

// Seg seguro
type Seg struct {
	InfResp InfResp `xml:"infResp"`
	InfSeg  InfSeg  `xml:"infSeg"`
	NApol   string  `xml:"nApol"`
	NAver   string  `xml:"nAver"`
}

// InfResp responsável pelo seguro
type InfResp struct {
	RespSeg string `xml:"respSeg"`
	CNPJ    string `xml:"CNPJ"`
	CPF     string `xml:"CPF"`
}

// InfSeg informações da seguradora
type InfSeg struct {
	XSeg string `xml:"xSeg"`
	CNPJ string `xml:"CNPJ"`
}

// ProdPred produto predominante
type ProdPred struct {
	TpCarga    string     `xml:"tpCarga"`
	XProd      string     `xml:"xProd"`
	InfLotacao InfLotacao `xml:"infLotacao"`
}

// InfLotacao informações de lotação
type InfLotacao struct {
	InfLocalCarrega    Local `xml:"infLocalCarrega"`
	InfLocalDescarrega Local `xml:"infLocalDescarrega"`
}

// Local local genérico
type Local struct {
	CEP string `xml:"CEP"`
}

// TotMDFe totalizadores
type TotMDFe struct {
	QCTe   string `xml:"qCTe"`
	QNFe   string `xml:"qNFe"`
	VCarga string `xml:"vCarga"`
	CUnid  string `xml:"cUnid"`
	QCarga string `xml:"qCarga"`
}

// InfAdic informações adicionais
type InfAdic struct {
	InfCpl string `xml:"infCpl"`
}

// InfRespTec responsável técnico
type InfRespTec struct {
	CNPJ     string `xml:"CNPJ"`
	XContato string `xml:"xContato"`
	Email    string `xml:"email"`
	Fone     string `xml:"fone"`
}

// InfCTeSupl suplemento do CT-e
type InfCTeSupl struct {
	QrCodCTe string `xml:"qrCodCTe"`
}

// InfMDFeSupl suplemento do MDF-e
type InfMDFeSupl struct {
	QrCodMDFe string `xml:"qrCodMDFe"`
}

// ProtMDFe protocolo do MDF-e
type ProtMDFe struct {
	InfProt InfProtMDFe `xml:"infProt"`
}

// InfProtMDFe informações do protocolo MDF-e
type InfProtMDFe struct {
	Id       string `xml:"Id,attr"`
	TpAmb    string `xml:"tpAmb"`
	VerAplic string `xml:"verAplic"`
	ChMDFe   string `xml:"chMDFe"`
	DhRecbto string `xml:"dhRecbto"`
	NProt    string `xml:"nProt"`
	DigVal   string `xml:"digVal"`
	CStat    string `xml:"cStat"`
	XMotivo  string `xml:"xMotivo"`
}

// Signature assinatura digital (simplificado)
type Signature struct {
	XMLName xml.Name `xml:"Signature"`
}

// ImpCTe impostos do CT-e
type ImpCTe struct {
	ICMS     ICMS   `xml:"ICMS"`
	VTotTrib string `xml:"vTotTrib"`
}

// ICMS imposto
type ICMS struct {
	ICMS00 *ICMS00 `xml:"ICMS00"`
	ICMS20 *ICMS20 `xml:"ICMS20"`
	ICMS45 *ICMS45 `xml:"ICMS45"`
	ICMS60 *ICMS60 `xml:"ICMS60"`
	ICMS90 *ICMS90 `xml:"ICMS90"`
}

// ICMS00 tributação normal
type ICMS00 struct {
	CST   string `xml:"CST"`
	VBC   string `xml:"vBC"`
	PICMS string `xml:"pICMS"`
	VICMS string `xml:"vICMS"`
}

// ICMS20 com redução de base
type ICMS20 struct {
	CST    string `xml:"CST"`
	PRedBC string `xml:"pRedBC"`
	VBC    string `xml:"vBC"`
	PICMS  string `xml:"pICMS"`
	VICMS  string `xml:"vICMS"`
}

// ICMS45 isento/não tributado
type ICMS45 struct {
	CST string `xml:"CST"`
}

// ICMS60 cobrado anteriormente
type ICMS60 struct {
	CST        string `xml:"CST"`
	VBCSTRet   string `xml:"vBCSTRet"`
	VICMSSTRet string `xml:"vICMSSTRet"`
	PICMSSTRet string `xml:"pICMSSTRet"`
	VCred      string `xml:"vCred"`
}

// ICMS90 outros
type ICMS90 struct {
	CST    string `xml:"CST"`
	PRedBC string `xml:"pRedBC"`
	VBC    string `xml:"vBC"`
	PICMS  string `xml:"pICMS"`
	VICMS  string `xml:"vICMS"`
	VCred  string `xml:"vCred"`
}

// InfModal informações do modal
type InfModal struct {
	VersaoModal string   `xml:"versaoModal,attr"`
	Rodo        *RodoCTe `xml:"rodo"`
}

// RodoCTe modal rodoviário do CT-e
type RodoCTe struct {
	RNTRC string `xml:"RNTRC"`
}

// Toma3 tomador do serviço
type Toma3 struct {
	Toma string `xml:"toma"`
}

// ============ Estruturas para Eventos ============

// ProcEventoCTe processo de evento CT-e
type ProcEventoCTe struct {
	XMLName      xml.Name     `xml:"procEventoCTe"`
	Versao       string       `xml:"versao,attr"`
	EventoCTe    EventoCTe    `xml:"eventoCTe"`
	RetEventoCTe RetEventoCTe `xml:"retEventoCTe"`
}

// EventoCTe evento do CT-e
type EventoCTe struct {
	Versao    string    `xml:"versao,attr"`
	InfEvento InfEvento `xml:"infEvento"`
	Signature Signature `xml:"Signature"`
}

// InfEvento informações do evento
type InfEvento struct {
	Id         string    `xml:"Id,attr"`
	COrgao     string    `xml:"cOrgao"`
	TpAmb      string    `xml:"tpAmb"`
	CNPJ       string    `xml:"CNPJ"`
	CPF        string    `xml:"CPF"`
	ChCTe      string    `xml:"chCTe"`
	ChMDFe     string    `xml:"chMDFe"`
	DhEvento   string    `xml:"dhEvento"`
	TpEvento   string    `xml:"tpEvento"`
	NSeqEvento string    `xml:"nSeqEvento"`
	DetEvento  DetEvento `xml:"detEvento"`
}

// DetEvento detalhes do evento
type DetEvento struct {
	VersaoEvento string      `xml:"versaoEvento,attr"`
	EvCancCTe    *EvCancCTe  `xml:"evCancCTe"`
	EvCancMDFe   *EvCancMDFe `xml:"evCancMDFe"`
}

// EvCancCTe cancelamento CT-e
type EvCancCTe struct {
	DescEvento string `xml:"descEvento"`
	NProt      string `xml:"nProt"`
	XJust      string `xml:"xJust"`
}

// EvCancMDFe cancelamento MDF-e
type EvCancMDFe struct {
	DescEvento string `xml:"descEvento"`
	NProt      string `xml:"nProt"`
	XJust      string `xml:"xJust"`
}

// RetEventoCTe retorno do evento
type RetEventoCTe struct {
	InfEvento InfEventoRet `xml:"infEvento"`
}

// InfEventoRet informações do retorno
type InfEventoRet struct {
	Id          string `xml:"Id,attr"`
	TpAmb       string `xml:"tpAmb"`
	VerAplic    string `xml:"verAplic"`
	COrgao      string `xml:"cOrgao"`
	CStat       string `xml:"cStat"`
	XMotivo     string `xml:"xMotivo"`
	ChCTe       string `xml:"chCTe"`
	ChMDFe      string `xml:"chMDFe"`
	TpEvento    string `xml:"tpEvento"`
	XEvento     string `xml:"xEvento"`
	NSeqEvento  string `xml:"nSeqEvento"`
	DhRegEvento string `xml:"dhRegEvento"`
	NProt       string `xml:"nProt"`
}

// ProcEventoMDFe processo de evento MDF-e
type ProcEventoMDFe struct {
	XMLName       xml.Name      `xml:"procEventoMDFe"`
	Versao        string        `xml:"versao,attr"`
	EventoMDFe    EventoMDFe    `xml:"eventoMDFe"`
	RetEventoMDFe RetEventoMDFe `xml:"retEventoMDFe"`
}

// EventoMDFe evento do MDF-e
type EventoMDFe struct {
	Versao    string    `xml:"versao,attr"`
	InfEvento InfEvento `xml:"infEvento"`
	Signature Signature `xml:"Signature"`
}

// RetEventoMDFe retorno do evento MDF-e
type RetEventoMDFe struct {
	Versao    string       `xml:"versao,attr"`
	InfEvento InfEventoRet `xml:"infEvento"`
}

// ParseDate converte string de data/hora do XML para time.Time
func ParseDate(dateStr string) (time.Time, error) {
	// Formato padrão: 2024-05-06T09:39:00-03:00
	return time.Parse(time.RFC3339, dateStr)
}

// ParseDateOnly converte string de data do XML para time.Time
func ParseDateOnly(dateStr string) (time.Time, error) {
	// Formato: 2024-05-06
	return time.Parse("2006-01-02", dateStr)
}
