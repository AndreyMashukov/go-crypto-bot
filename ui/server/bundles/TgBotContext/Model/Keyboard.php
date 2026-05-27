<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

use TgBotApi\BotApiBase\Exception\BadArgumentException;
use TgBotApi\BotApiBase\Type\InlineKeyboardButtonType as IKButton;
use TgBotApi\BotApiBase\Type\InlineKeyboardMarkupType;
use TgBotApi\BotApiBase\Type\KeyboardButtonType as KButton;
use TgBotApi\BotApiBase\Type\ReplyKeyboardMarkupType;

class Keyboard
{
    private array $actions = [];

    private array $inlineActions = [];

    private bool $oneRow = false;

    private int $rowLimit = 3;

    /**
     * @param string $action
     *
     * @throws BadArgumentException
     *
     * @return $this
     */
    public function addAction(string $action): self
    {
        $this->actions[] = [KButton::create($action)];

        return $this;
    }

    /**
     * @param string $action
     *
     * @throws BadArgumentException
     *
     * @return $this
     */
    public function addContactRequest(string $action): self
    {
        $button                 = KButton::create($action);
        $button->requestContact = true;
        $this->actions[]        = [$button];

        return $this;
    }

    /**
     * @param string $action
     * @param string $url
     * @param string $data
     *
     * @throws BadArgumentException
     *
     * @return $this
     */
    public function addInlineAction(string $action, string $url = null, string $data = null): self
    {
        $this->inlineActions[] = [IKButton::create($action, [
            'url'          => $url,
            'callbackData' => $data,
        ])];

        return $this;
    }

    /**
     * @throws BadArgumentException
     *
     * @return ReplyKeyboardMarkupType
     */
    public function getKeyboardMarkup(): ReplyKeyboardMarkupType
    {
        return ReplyKeyboardMarkupType::create($this->actions, [
            'oneTimeKeyboard' => true,
            'resizeKeyboard'  => true,
        ]);
    }

    /**
     * @return InlineKeyboardMarkupType
     */
    public function getInlineKeyboardMarkup(): InlineKeyboardMarkupType
    {
        if ($this->oneRow) {
            $buttons = [];
            $index   = 0;

            foreach ($this->inlineActions as $button) {
                $buttons[$index][] = array_shift($button);

                if (\count($buttons[$index]) >= $this->rowLimit) {
                    ++$index;
                }
            }

            return InlineKeyboardMarkupType::create($buttons);
        }

        return InlineKeyboardMarkupType::create($this->inlineActions);
    }

    /**
     * @throws BadArgumentException
     *
     * @return InlineKeyboardMarkupType|ReplyKeyboardMarkupType
     */
    public function getKeyboard()
    {
        if (\count($this->inlineActions) > 0) {
            return $this->getInlineKeyboardMarkup();
        }

        return $this->getKeyboardMarkup();
    }

    public function toOneRow(int $limit = 3): self
    {
        $this->oneRow   = true;
        $this->rowLimit = $limit;

        return $this;
    }
}
