package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Rivaldz/my-template-go/internal/controller/restapi/v1/request"
	"github.com/Rivaldz/my-template-go/internal/controller/restapi/v1/response"
	"github.com/Rivaldz/my-template-go/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     Create comment
// @Description Add a comment to a post
// @ID          create-comment
// @Tags        comments
// @Accept      json
// @Produce     json
// @Param       postID  path     string                true "Post ID"
// @Param       request body     request.CreateComment true "Comment data"
// @Success     201      {object} entity.Comment
// @Failure     400      {object} response.Error
// @Failure     401      {object} response.Error
// @Failure     404      {object} response.Error
// @Failure     500      {object} response.Error
// @Security    BearerAuth
// @Router      /posts/{postID}/comments [post]
func (r *V1) createComment(ctx *fiber.Ctx) error {
	userID, ok := authUserID(ctx)
	if !ok {
		return unauthorizedResponse(ctx)
	}

	postID := ctx.Params("postID")

	var body request.CreateComment

	if !r.bindBody(ctx, "createComment", &body) {
		return badRequestResponse(ctx)
	}

	comment, err := r.c.Create(ctx.UserContext(), userID, postID, body.Body)
	if err != nil {
		r.l.Error(err, "restapi - v1 - createComment")

		if errors.Is(err, entity.ErrPostNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "post not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusCreated).JSON(comment)
}

// @Summary     List comments
// @Description List comments for a post
// @ID          list-comments
// @Tags        comments
// @Produce     json
// @Param       postID path    string true "Post ID"
// @Param       limit  query   int    false "Limit"  default(10)
// @Param       offset query   int    false "Offset" default(0)
// @Success     200    {object} response.CommentList
// @Failure     401    {object} response.Error
// @Failure     500    {object} response.Error
// @Security    BearerAuth
// @Router      /posts/{postID}/comments [get]
func (r *V1) listComments(ctx *fiber.Ctx) error {
	postID := ctx.Params("postID")

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(ctx.Query("offset", "0"))
	if err != nil {
		offset = 0
	}

	comments, total, err := r.c.ListByPost(ctx.UserContext(), postID, limit, offset)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listComments")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(response.CommentList{
		Comments: comments,
		Total:    total,
	})
}

// @Summary     Get comment
// @Description Get a comment by ID
// @ID          get-comment
// @Tags        comments
// @Produce     json
// @Param       id  path     string true "Comment ID"
// @Success     200 {object} entity.Comment
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /comments/{id} [get]
func (r *V1) getComment(ctx *fiber.Ctx) error {
	commentID := ctx.Params("id")

	comment, err := r.c.Get(ctx.UserContext(), commentID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getComment")

		if errors.Is(err, entity.ErrCommentNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "comment not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(comment)
}

// @Summary     Update comment
// @Description Update a comment (author only)
// @ID          update-comment
// @Tags        comments
// @Accept      json
// @Produce     json
// @Param       id      path     string                true "Comment ID"
// @Param       request body     request.UpdateComment true "Updated comment data"
// @Success     200     {object} entity.Comment
// @Failure     400     {object} response.Error
// @Failure     401     {object} response.Error
// @Failure     403     {object} response.Error
// @Failure     404     {object} response.Error
// @Failure     500     {object} response.Error
// @Security    BearerAuth
// @Router      /comments/{id} [put]
func (r *V1) updateComment(ctx *fiber.Ctx) error {
	userID, ok := authUserID(ctx)
	if !ok {
		return unauthorizedResponse(ctx)
	}

	commentID := ctx.Params("id")

	var body request.UpdateComment

	if !r.bindBody(ctx, "updateComment", &body) {
		return badRequestResponse(ctx)
	}

	comment, err := r.c.Update(ctx.UserContext(), userID, commentID, body.Body)
	if err != nil {
		r.l.Error(err, "restapi - v1 - updateComment")

		if errors.Is(err, entity.ErrCommentNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "comment not found")
		}

		if errors.Is(err, entity.ErrCommentForbidden) {
			return errorResponse(ctx, http.StatusForbidden, "forbidden")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(comment)
}

// @Summary     Delete comment
// @Description Delete a comment (author only)
// @ID          delete-comment
// @Tags        comments
// @Param       id  path     string true "Comment ID"
// @Success     204 "No Content"
// @Failure     401 {object} response.Error
// @Failure     403 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /comments/{id} [delete]
func (r *V1) deleteComment(ctx *fiber.Ctx) error { //nolint:dupl // mirrors the standard delete handler template
	userID, ok := authUserID(ctx)
	if !ok {
		return unauthorizedResponse(ctx)
	}

	err := r.c.Delete(ctx.UserContext(), userID, ctx.Params("id"))
	if err != nil {
		r.l.Error(err, "restapi - v1 - deleteComment")

		if errors.Is(err, entity.ErrCommentNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "comment not found")
		}

		if errors.Is(err, entity.ErrCommentForbidden) {
			return errorResponse(ctx, http.StatusForbidden, "forbidden")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.SendStatus(http.StatusNoContent)
}
