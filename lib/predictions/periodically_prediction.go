package predictions

import ()

func Periodically_Prediction(w int, b int, s int, tw int) int {
	//ここでは1,2時間後の負荷を取得する
	//引数は台数情報と重み情報
	//返り値は1時間後に必要な重み情報
	//スクリプトに渡す値は予測開始時間，予測終了時間,2時間後の時間
	//スクリプトから貰う値は1時間後の負荷と2時間後の負荷
	//前回値を取得した時から1時間経ったら新たにスクリプトを実行する
	//それまでは前回値を返す

	return 0
}
